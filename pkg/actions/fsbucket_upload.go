package actions

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/jlaffaye/ftp"
	"github.com/miton18/helper/set"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

//go:embed fsbucket_upload_doc.md
var actionFSBucketUploadDoc string

const (
	SchemeFile  = "file"
	SchemeHTTP  = "http"
	SchemeHTTPS = "https"
)

var allowedScheme = set.New(SchemeFile, SchemeHTTP, SchemeHTTPS)

type (
	ActionFSBucketUpload struct {
		provider.Provider
	}

	fsbucketUpload struct {
		FSBucketID types.String `tfsdk:"fsbucket_id"`
		Files      types.List   `tfsdk:"file"`
	}
)

func FSBucketUpload() action.Action {
	return &ActionFSBucketUpload{}
}

func (r fsbucketUpload) GetFiles(ctx context.Context, diags *diag.Diagnostics) []fileUpload {
	var fileBlocks []fileUpload
	diags.Append(r.Files.ElementsAs(ctx, &fileBlocks, false)...)
	return fileBlocks
}

type fileUpload struct {
	LocalPath  types.String `tfsdk:"local_path"`
	RemotePath types.String `tfsdk:"remote_path"`
}

func (f *fileUpload) GetLocal(diags *diag.Diagnostics) *url.URL {
	fileURL, err := url.Parse(f.LocalPath.ValueString())
	if err != nil {
		diags.AddError("failed to parse local path as URL", err.Error())
	}
	return fileURL
}

// resolvedFile represents a file ready to be uploaded
type resolvedFile struct {
	Content    io.ReadCloser
	RemotePath string
	SourceDesc string // Description for logging/progress
}

func (a *ActionFSBucketUpload) Configure(ctx context.Context, req action.ConfigureRequest, res *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	if provider, ok := req.ProviderData.(provider.Provider); ok {
		a.Provider = provider
	}

	tflog.Debug(ctx, "Configured", map[string]any{"org": a.Organization()})
}

func (a *ActionFSBucketUpload) Metadata(ctx context.Context, req action.MetadataRequest, res *action.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_fsbucket_upload"
}

func (a *ActionFSBucketUpload) Schema(ctx context.Context, req action.SchemaRequest, res *action.SchemaResponse) {
	res.Schema = schema.Schema{
		MarkdownDescription: actionFSBucketUploadDoc,
		Attributes: map[string]schema.Attribute{
			"fsbucket_id": schema.StringAttribute{
				Required:    true,
				Description: "FSBucket ID to upload to",
				Validators: []validator.String{
					pkg.NewStringValidator(
						"must be an addon ID",
						func(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
							if req.ConfigValue.IsNull() {
								res.Diagnostics.AddError("cannot be null", "fsbucket_id is null")
							} else if req.ConfigValue.IsUnknown() {
								return
							}

							if !strings.HasPrefix(req.ConfigValue.ValueString(), "bucket_") {
								res.Diagnostics.AddError("expect a valid fsbucket ID", "ID doesn't start with 'bucket_'")
							}
						},
					),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"file": schema.ListNestedBlock{
				MarkdownDescription: "File or directory to upload",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"local_path": schema.StringAttribute{
							Required:    true,
							Description: "Local file or directory path to upload. Supports file://, http://, and https:// schemes.",
							Validators: []validator.String{
								pkg.NewStringValidator("file with scheme prefix", func(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
									if req.ConfigValue.IsUnknown() {
										return
									}

									value := req.ConfigValue.ValueString()
									fileURL, err := url.Parse(value)
									if err != nil {
										res.Diagnostics.AddError("expect a valid URL", err.Error())
										return
									}

									if !allowedScheme.Contains(fileURL.Scheme) {
										res.Diagnostics.AddError(
											"unsupported URL scheme",
											fmt.Sprintf(
												"supported schemes are: %s",
												strings.Join(allowedScheme.Slice(), ", "),
											),
										)
									}
								}),
							},
						},
						"remote_path": schema.StringAttribute{
							Required:    true,
							Description: "Remote path in the FSBucket (destination)",
						},
					},
				},
			},
		},
	}
}

func (a *ActionFSBucketUpload) Invoke(ctx context.Context, req action.InvokeRequest, res *action.InvokeResponse) {
	tflog.Debug(ctx, "Invoke fsbucket_upload", map[string]any{"config": req.Config})

	cfg := helper.From[fsbucketUpload](ctx, req.Config, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	files := cfg.GetFiles(ctx, &res.Diagnostics)

	if len(files) == 0 {
		res.Diagnostics.AddWarning("no files specified", "skipping action")
		return
	}

	Progress(res, "Retrieving fsbucket credentials...")
	fsbucketEnvRes := tmp.GetAddonEnv(ctx, a.Client(), a.Organization(), cfg.FSBucketID.ValueString())
	if fsbucketEnvRes.HasError() {
		res.Diagnostics.AddError("failed to get FSBucket credentials", fsbucketEnvRes.Error().Error())
		return
	}
	fsbucketEnv := fsbucketEnvRes.Payload()

	envMap := pkg.Reduce(*fsbucketEnv, map[string]string{}, func(m map[string]string, v tmp.EnvVar) map[string]string {
		m[v.Name] = v.Value
		return m
	})

	host := envMap["BUCKET_HOST"]
	username := envMap["BUCKET_FTP_USERNAME"]
	password := envMap["BUCKET_FTP_PASSWORD"]

	if host == "" || username == "" || password == "" {
		res.Diagnostics.AddError("missing FTP credentials", "BUCKET_HOST, BUCKET_FTP_USERNAME, or BUCKET_FTP_PASSWORD not found")
		return
	}

	Progress(res, "Connecting to FSBucket FTP...")
	conn, err := ftp.Dial(host+":21", ftp.DialWithContext(ctx), ftp.DialWithTimeout(30*time.Second), ftp.DialWithDebugOutput(os.Stdout))
	if err != nil {
		res.Diagnostics.AddError("failed to connect to FTP server", err.Error())
		return
	}
	defer func() {
		if err := conn.Quit(); err != nil {
			res.Diagnostics.AddWarning("failed to close FTP connection", err.Error())
		}
	}()

	Progress(res, "Login to FTP server...")

	if err := loginWithRetry(ctx, conn, username, password, res); err != nil {
		res.Diagnostics.AddError("failed to login to FTP server after 3 attempts", err.Error())
		return
	}

	uploadCount := 0
	for i, fileBlock := range files {
		remotePath := fileBlock.RemotePath.ValueString()
		localURL := fileBlock.GetLocal(&res.Diagnostics)
		if res.Diagnostics.HasError() {
			continue
		}

		tflog.Info(ctx, "Processing file block", map[string]any{
			"index":       i,
			"local_url":   localURL.String(),
			"scheme":      localURL.Scheme,
			"remote_path": remotePath,
		})

		var resolvedFiles []resolvedFile
		var err error

		switch localURL.Scheme {
		case SchemeFile:
			resolvedFiles, err = resolveFileSource(ctx, localURL, remotePath)
		case SchemeHTTP, SchemeHTTPS:
			resolvedFiles, err = resolveHTTPSource(ctx, localURL, remotePath)
		default:
			res.Diagnostics.AddError(
				"unsupported URL scheme",
				fmt.Sprintf("scheme '%s' is not supported", localURL.Scheme),
			)
			continue
		}
		if err != nil {
			res.Diagnostics.AddError(
				fmt.Sprintf("failed to resolve source '%s'", localURL.String()),
				err.Error(),
			)
			continue
		}

		// Upload all resolved files
		for i, resolved := range resolvedFiles {
			Progress(
				res,
				"Uploading '%s' -> '%s' (%d/%d)",
				resolved.SourceDesc, resolved.RemotePath, i+1, len(resolvedFiles),
			)

			err := uploadFileFromReader(ctx, conn, resolved.Content, resolved.RemotePath)
			if err != nil {
				res.Diagnostics.AddError(
					fmt.Sprintf("failed to upload '%s'", resolved.SourceDesc),
					err.Error(),
				)
				continue
			}
			uploadCount++
			Progress(res, "Uploaded '%s'", resolved.SourceDesc)
		}
	}

	Progress(res, "Successfully uploaded %d file(s)", uploadCount)
}

// resolveFileSource resolves file:// URLs to one or more files to upload
func resolveFileSource(ctx context.Context, localURL *url.URL, remotePath string) ([]resolvedFile, error) {
	localPath := localURL.Host + localURL.Path

	// Resolve to absolute path
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for '%s': %w", localPath, err)
	}
	localPath = absPath

	info, err := os.Stat(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to access local path '%s': %w", localPath, err)
	}

	if info.IsDir() {
		return resolveDirectory(ctx, localPath, remotePath)
	}

	file, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file '%s': %w", localPath, err)
	}

	return []resolvedFile{{
		Content:    file,
		RemotePath: remotePath,
		SourceDesc: localPath,
	}}, nil
}

// resolveHTTPSource resolves http(s):// URLs to a single file to upload
func resolveHTTPSource(ctx context.Context, localURL *url.URL, remotePath string) ([]resolvedFile, error) {
	content, err := downloadHTTPContent(ctx, localURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve HTTP ressource")
	}

	return []resolvedFile{{
		Content:    content,
		RemotePath: remotePath,
		SourceDesc: localURL.String(),
	}}, nil
}

// resolveDirectory resolves a directory to multiple files
func resolveDirectory(ctx context.Context, localDir, remoteDir string) ([]resolvedFile, error) {
	var files []resolvedFile

	err := filepath.WalkDir(localDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			subFiles, err := resolveDirectory(ctx, path, remoteDir)
			if err != nil {
				return err
			}
			files = append(files, subFiles...)
		}

		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		relPath = filepath.ToSlash(relPath)
		remotePath := filepath.ToSlash(filepath.Join(remoteDir, relPath))

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file '%s': %w", path, err)
		}

		files = append(files, resolvedFile{
			Content:    file,
			RemotePath: remotePath,
			SourceDesc: path,
		})

		return nil
	})
	if err != nil {
		for _, f := range files {
			f.Content.Close()
		}
		return nil, err
	}

	return files, nil
}

// downloadHTTPContent downloads content from an HTTP(S) URL
func downloadHTTPContent(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	return resp.Body, nil
}

// loginWithRetry attempts to login to FTP server with retry logic
// FTP accounts might not be available immediately after creation
func loginWithRetry(ctx context.Context, conn *ftp.ServerConn, username, password string, res *action.InvokeResponse) error {
	return retry(ctx, 3, 4*time.Second, func() error {
		return conn.Login(username, password)
	}, func(attempt int, err error) {
		tflog.Warn(ctx, "FTP login failed, retrying...", map[string]any{
			"attempt": attempt,
			"error":   err.Error(),
		})
		Progress(res, "FTP login failed (attempt %d/3), retrying in 4s...", attempt)
	})
}

func uploadFileFromReader(ctx context.Context, conn *ftp.ServerConn, content io.ReadCloser, remotePath string) error {
	defer func() {
		if err := content.Close(); err != nil {
			tflog.Warn(ctx, "failed to close file reader", map[string]any{"err": err.Error()})
		}
	}()
	remoteDir := filepath.Dir(remotePath)
	if remoteDir != "." && remoteDir != "/" {
		if err := createRemoteDir(conn, remoteDir); err != nil {
			return fmt.Errorf("failed to create remote directory: %w", err)
		}
	}

	err := conn.Stor(remotePath, content)
	if err != nil {
		return fmt.Errorf("failed to store file: %w", err)
	}

	return nil
}

func createRemoteDir(conn *ftp.ServerConn, path string) error {
	path = filepath.ToSlash(filepath.Clean(path))
	if path == "." || path == "/" {
		return nil
	}

	parts := strings.Split(path, "/")
	currentPath := ""

	for _, part := range parts {
		if part == "" {
			continue
		}

		if currentPath == "" {
			currentPath = part
		} else {
			currentPath = currentPath + "/" + part
		}

		err := conn.MakeDir(currentPath)
		if err != nil {
			if err := conn.ChangeDir(currentPath); err != nil {
				return fmt.Errorf("failed to create or access directory %s: %w", currentPath, err)
			}
			_ = conn.ChangeDir("/")
		}
	}

	return nil
}

// retry executes a function with retry logic
// maxAttempts: maximum number of attempts
// interval: delay between retries
// fn: function to execute
// onRetry: optional callback called before each retry (not called on first attempt or after last failure)
func retry(_ context.Context, maxAttempts int, interval time.Duration, fn func() error, onRetry func(attempt int, err error)) error {
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Don't sleep or call onRetry after last attempt
		if attempt < maxAttempts {
			if onRetry != nil {
				onRetry(attempt, lastErr)
			}
			time.Sleep(interval)
		}
	}

	return lastErr
}
