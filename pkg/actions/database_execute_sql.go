package actions

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	pgx "github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func ExecuteDatabaseSQL() action.Action {
	return &ActionExecuteDatabaseSQL{}
}

type ActionExecuteDatabaseSQL struct {
	provider.Provider
}

type DatabaseQuery struct {
	DatabaseID types.String `tfsdk:"database_id"`
	Query      types.String `tfsdk:"query"`
	OutputJSON types.String `tfsdk:"output_json"`
}

func (a *ActionExecuteDatabaseSQL) Configure(ctx context.Context, req action.ConfigureRequest, res *action.ConfigureResponse) {
	tflog.Debug(ctx, "Configure()")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	if provider, ok := req.ProviderData.(provider.Provider); ok {
		a.Provider = provider
	}

	tflog.Debug(ctx, "Configured", map[string]any{"org": a.Organization()})
}

//go:embed database_execute_sql_doc.md
var actionExecuteDatabaseSQLDoc string

func (a *ActionExecuteDatabaseSQL) Schema(ctx context.Context, req action.SchemaRequest, res *action.SchemaResponse) {
	res.Schema = schema.Schema{
		MarkdownDescription: actionExecuteDatabaseSQLDoc,
		Attributes: map[string]schema.Attribute{
			"database_id": schema.StringAttribute{
				Required:    true,
				Description: "Database ID to execute the query on",
				Validators: []validator.String{
					pkg.NewStringValidator(
						"must be a database addon ID",
						func(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
							if req.ConfigValue.IsNull() {
								res.Diagnostics.AddError("cannot be null", "database_id is null")
							} else if req.ConfigValue.IsUnknown() {
								return
							}

							if !strings.HasPrefix(req.ConfigValue.ValueString(), "postgresql") {
								res.Diagnostics.AddError("expect a valid addon ID", "ID doesn't start with 'postgres_'")
							}
						},
					),
				},
			},
			"query": schema.StringAttribute{Required: true, Description: "SQL query to execute"},
			"output_json": schema.StringAttribute{
				Optional:    true,
				Description: "file path and name to write query result into, starts with file://",
				Validators: []validator.String{pkg.NewStringValidator("must starts with file://", func(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
					if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
						return
					}

					if !strings.HasPrefix(req.ConfigValue.ValueString(), "file://") {
						res.Diagnostics.AddAttributeError(req.Path, "expect file:// as prefix", "unexpected prefix")
					}
				})},
			},
		},
	}
}

func (a *ActionExecuteDatabaseSQL) Metadata(ctx context.Context, req action.MetadataRequest, res *action.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_database_query"
}

//func (a *ActionExecuteDatabaseSQL) ValidateConfig(ctx context.Context, req action.ValidateConfigRequest, res *action.ValidateConfigResponse) {}

func (a *ActionExecuteDatabaseSQL) Invoke(ctx context.Context, req action.InvokeRequest, res *action.InvokeResponse) {
	cfg := helper.From[DatabaseQuery](ctx, req.Config, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Invoke database_execute_sql", map[string]any{"config": req.Config})

	switch {
	case strings.HasPrefix(cfg.DatabaseID.ValueString(), "postgresql_"):
		a.InvokePG(ctx, cfg, ProgressWrapper(res), &res.Diagnostics)
	}
}

func (a *ActionExecuteDatabaseSQL) InvokePG(ctx context.Context, cfg *DatabaseQuery, progress func(msg string, args ...any), diags *diag.Diagnostics) {
	progress("Resolving database ID")

	addonID, err := tmp.RealIDToAddonID(ctx, a.Client(), a.Organization(), cfg.DatabaseID.ValueString())
	if err != nil {
		diags.AddError("failed to resolve database ID", err.Error())
		return
	}

	progress("Fetching database credentials")
	pgRes := tmp.GetPostgreSQL(ctx, a.Client(), addonID)
	if pgRes.HasError() {
		diags.AddError("failed to get database credentials", pgRes.Error().Error())
		return
	}
	pg := pgRes.Payload()

	query := cfg.Query.ValueString()
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", pg.User, pg.Password, pg.Host, pg.Port, pg.Database)

	progress("Opening database connection")
	conn, err := Connect(ctx, dsn)
	if err != nil {
		diags.AddError("failed to connect to databse", err.Error())
		return
	}
	defer func() {
		progress("Closing database connection")
		if err := conn.Close(ctx); err != nil {
			diags.AddWarning("failed to close database connection", err.Error())
		}
	}()

	if cfg.OutputJSON.IsNull() || cfg.OutputJSON.IsUnknown() {
		// no output, multiple statements allowed
		progress("Executing database query")
		result, err := conn.Exec(ctx, query)
		if err != nil {
			diags.AddError("failed execute query", err.Error())
			return
		}
		tflog.Debug(ctx, "executed statement", map[string]any{
			"rows": result.RowsAffected(),
			"str":  result.String(),
		})
		return
	} else {
		progress("Executing database query")
		result, err := conn.Query(ctx, query)
		if err != nil {
			diags.AddError("failed execute query", err.Error())
			return
		}
		defer result.Close()

		progress("Serializing databse query results")
		data, err := PgSqlRowsToJson(result)
		if err != nil {
			diags.AddError("failed to serialize result", err.Error())
			return
		}

		progress("Writing databse query results")
		file, err := os.OpenFile(
			strings.TrimPrefix(cfg.OutputJSON.ValueString(), "file://"),
			os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm,
		)
		if err != nil {
			diags.AddError("failed to open output file", err.Error())
		}
		defer func() {
			if err := file.Close(); err != nil {
				diags.AddWarning("failed to close output file", err.Error())
			}
		}()

		_, err = file.Write(data)
		if err != nil {
			diags.AddError("failed to write result", err.Error())
			return
		}

		if err := file.Sync(); err != nil {
			diags.AddError("failed to sync file", err.Error())
			return
		}
	}
}

// Connect attempts to establish a database connection with retry logic.
// It retries every second until the connection succeeds or the context expires.
// Returns the last error encountered if the context expires.
func Connect(ctx context.Context, dsn string) (*pgx.Conn, error) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastErr error
	var conn *pgx.Conn
	i := 0

	conn, lastErr = pgx.Connect(ctx, dsn)
	if lastErr == nil {
		return conn, nil
	}
	i++

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context expired: %w", lastErr)
		case <-ticker.C:
			conn, lastErr = pgx.Connect(ctx, dsn)
			if lastErr == nil {
				return conn, nil
			}
			tflog.Debug(ctx, "retrying connection to PG", map[string]any{
				"error": lastErr.Error(),
				"retry": i,
			})
			i++
		}
	}
}

func PgSqlRowsToJson(rows pgx.Rows) ([]byte, error) {
	fieldDescriptions := rows.FieldDescriptions()
	var columns []string
	for _, col := range fieldDescriptions {
		columns = append(columns, col.Name)
	}

	tableData := make([]map[string]any, 0)

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to read row values: %w", err)
		}

		entry := make(map[string]any)
		for i, col := range columns {
			entry[col] = values[i]
		}
		tableData = append(tableData, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return json.Marshal(tableData)
}
