package application

import (
	"bufio"
	"compress/zlib"
	"context"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/sideband"
	"github.com/go-git/go-git/v5/plumbing/revlist"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
)

// streamPush force-pushes target to the master branch of the remote.
//
// go-git's Remote.Push lists every object of the pack in memory (decoded
// when smaller than 16 KiB, plus the delta search) before writing it, so its
// memory grows with the repository size (#469). This push negotiates the
// references the same way but writes the pack one object at a time: only the
// list of hashes to send is kept in memory.
//
// It returns false when the remote branch already points to target.
func streamPush(
	ctx context.Context,
	s storer.EncodedObjectStorer,
	remoteURL string,
	auth transport.AuthMethod,
	target plumbing.Hash,
	progress io.Writer,
) (bool, error) {
	ep, err := transport.NewEndpoint(remoteURL)
	if err != nil {
		return false, err
	}
	c, err := client.NewClient(ep)
	if err != nil {
		return false, err
	}
	sess, err := c.NewReceivePackSession(ep, auth)
	if err != nil {
		return false, err
	}
	defer sess.Close()

	ar, err := sess.AdvertisedReferencesContext(ctx)
	if err != nil {
		return false, err
	}
	remoteRefs, err := ar.AllReferences()
	if err != nil {
		return false, err
	}

	// force push: the update is made against the current remote value,
	// whatever it is
	old := plumbing.ZeroHash
	if ref, ok := remoteRefs[plumbing.Master]; ok {
		old = ref.Hash()
	}
	if old == target {
		return false, nil
	}

	req := packp.NewReferenceUpdateRequestFromCapabilities(ar.Capabilities)
	req.Commands = []*packp.Command{{Name: plumbing.Master, Old: old, New: target}}
	if progress != nil {
		req.Progress = progress
		if ar.Capabilities.Supports(capability.Sideband64k) {
			_ = req.Capabilities.Set(capability.Sideband64k)
		} else if ar.Capabilities.Supports(capability.Sideband) {
			_ = req.Capabilities.Set(capability.Sideband)
		}
	}

	// objects the remote already has are not sent
	haves := []plumbing.Hash{}
	for _, ref := range remoteRefs {
		if ref.Type() == plumbing.HashReference {
			haves = append(haves, ref.Hash())
		}
	}
	hashes, err := revlist.Objects(s, []plumbing.Hash{target}, haves)
	if err != nil {
		return false, fmt.Errorf("failed to list objects to push: %w", err)
	}

	rd, wr := io.Pipe()
	req.Packfile = rd
	done := make(chan error, 1)
	go func() {
		done <- wr.CloseWithError(writePack(wr, s, hashes))
	}()

	var rs *packp.ReportStatus
	if ep.Protocol == "http" || ep.Protocol == "https" {
		rs, err = httpReceivePack(ctx, ep, auth, req)
	} else {
		rs, err = sess.ReceivePack(ctx, req)
	}
	if err != nil {
		// unblock the pack writer
		_ = rd.Close()
		<-done //nolint:errcheck // the writer only fails because the read side was closed
		return false, err
	}
	if err := <-done; err != nil {
		return false, err
	}
	if rs != nil {
		if err := rs.Error(); err != nil {
			return false, err
		}
	}

	return true, nil
}

// httpReceivePack sends the receive-pack request with a streamed body: go-git's
// HTTP transport encodes the whole request, packfile included, in a memory
// buffer before sending it.
func httpReceivePack(ctx context.Context, ep *transport.Endpoint, auth transport.AuthMethod, req *packp.ReferenceUpdateRequest) (*packp.ReportStatus, error) {
	body, bodyWriter := io.Pipe()
	go func() {
		bodyWriter.CloseWithError(req.Encode(bodyWriter))
	}()
	defer body.Close()

	url := fmt.Sprintf("%s/%s", ep.String(), transport.ReceivePackServiceName)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("User-Agent", capability.DefaultAgent())
	httpReq.Header.Set("Accept", "application/x-git-receive-pack-result")
	httpReq.Header.Set("Content-Type", "application/x-git-receive-pack-request")
	if a, ok := auth.(interface{ SetAuth(*http.Request) }); ok {
		a.SetAuth(httpReq)
	}

	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(res.Body, 4<<10))
		return nil, &receivePackError{
			StatusCode: res.StatusCode,
			Status:     res.Status,
			Body:       strings.TrimSpace(string(msg)),
		}
	}

	var r io.Reader = res.Body
	if req.Capabilities.Supports(capability.Sideband64k) {
		d := sideband.NewDemuxer(sideband.Sideband64k, r)
		d.Progress = req.Progress
		r = d
	} else if req.Capabilities.Supports(capability.Sideband) {
		d := sideband.NewDemuxer(sideband.Sideband, r)
		d.Progress = req.Progress
		r = d
	}

	if !req.Capabilities.Supports(capability.ReportStatus) {
		return nil, nil
	}
	report := packp.NewReportStatus()
	if err := report.Decode(r); err != nil {
		return nil, fmt.Errorf("receive-pack: %w", err)
	}
	return report, nil
}

// receivePackError is an HTTP error answered by the remote to the push
type receivePackError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *receivePackError) Error() string {
	return fmt.Sprintf("receive-pack: %s: %s", e.Status, e.Body)
}

// tooLarge tells whether the remote rejected the push because of its size
func (e *receivePackError) tooLarge() bool {
	return e.StatusCode == http.StatusRequestEntityTooLarge
}

// writePack writes a packfile (version 2, no deltas) holding the given
// objects, reading and compressing them one at a time.
func writePack(w io.Writer, s storer.EncodedObjectStorer, hashes []plumbing.Hash) error {
	bw := bufio.NewWriterSize(w, 64<<10)
	checksum := sha1.New()
	out := io.MultiWriter(bw, checksum)

	header := make([]byte, 12)
	copy(header, "PACK")
	binary.BigEndian.PutUint32(header[4:], 2)
	binary.BigEndian.PutUint32(header[8:], uint32(len(hashes)))
	if _, err := out.Write(header); err != nil {
		return err
	}

	zw := zlib.NewWriter(out)
	for _, h := range hashes {
		if err := writePackEntry(out, zw, s, h); err != nil {
			return fmt.Errorf("object %s: %w", h, err)
		}
	}

	if _, err := bw.Write(checksum.Sum(nil)); err != nil {
		return err
	}
	return bw.Flush()
}

func writePackEntry(out io.Writer, zw *zlib.Writer, s storer.EncodedObjectStorer, h plumbing.Hash) error {
	obj, err := s.EncodedObject(plumbing.AnyObject, h)
	if err != nil {
		return err
	}

	if _, err := out.Write(packEntryHeader(obj.Type(), obj.Size())); err != nil {
		return err
	}

	r, err := obj.Reader()
	if err != nil {
		return err
	}
	defer r.Close()

	zw.Reset(out)
	if _, err := io.Copy(zw, r); err != nil {
		return err
	}
	return zw.Close()
}

// packEntryHeader encodes the type and the inflated size of a pack entry:
// type in bits 4-6 of the first byte, size on the remaining bits, 7 bits per
// byte with the high bit telling whether another byte follows.
func packEntryHeader(t plumbing.ObjectType, size int64) []byte {
	b := byte(t)<<4 | byte(size&0x0f)
	size >>= 4
	header := []byte{}
	for size != 0 {
		header = append(header, b|0x80)
		b = byte(size & 0x7f)
		size >>= 7
	}
	return append(header, b)
}
