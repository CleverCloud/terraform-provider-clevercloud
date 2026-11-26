package pkg

import (
	"context"
	"sync"

	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk/models"
	"go.clever-cloud.dev/sdk/services/products"
)

// AddonProviders lists the add-on providers of the catalogue and their plans.
// The catalogue is downloaded once per client, see cachedCatalogue.
func AddonProviders(ctx context.Context, cc *client.Client) client.Response[[]models.AddonProviderInfoFullView] {
	return cachedCatalogue(cc, "addonproviders", func() client.Response[[]models.AddonProviderInfoFullView] {
		return SDK(cc).V2().Products().Addonproviders().ListProductAddons(ctx)
	})
}

// ProductInstances lists the application runtimes, those available to ownerID
// when it is set. The catalogue is downloaded once per client and owner, see
// cachedCatalogue.
func ProductInstances(ctx context.Context, cc *client.Client, ownerID *string) client.Response[[]models.AvailableInstanceView] {
	name := "instances"
	var opts []products.Option
	if ownerID != nil {
		name += "?for=" + *ownerID
		opts = append(opts, products.WithForparam(*ownerID))
	}

	return cachedCatalogue(cc, name, func() client.Response[[]models.AvailableInstanceView] {
		return SDK(cc).V2().Products().Instances().ListProductRuntimes(ctx, opts...)
	})
}

type catalogueKey struct {
	cc   *client.Client
	name string
}

// catalogueEntry holds the answer of one catalogue route. Its mutex is held
// for the whole download: the callers arriving meanwhile wait for it, then
// read the answer instead of downloading it again.
type catalogueEntry struct {
	mu       sync.Mutex
	response any
}

var catalogues sync.Map // catalogueKey -> *catalogueEntry

// cachedCatalogue downloads a product catalogue once per client and name, and
// answers from memory afterwards.
//
// Every resource reads a catalogue before it creates anything, so an apply
// creating twenty add-ons downloaded the same 200 KB document twenty times, ten
// at once with the default parallelism. Besides the wasted time, the proxy in
// front of the API ends an HTTP/2 connection carrying several such downloads at
// once (GOAWAY, ENHANCE_YOUR_CALM), which fails every request in flight on it.
//
// A catalogue does not change during a plan or an apply, which is how long a
// provider process lives, so the answer is never refreshed. A failed download
// is not kept: the next caller tries again.
//
// The answer is shared between the callers, they must not modify it.
func cachedCatalogue[T any](cc *client.Client, name string, download func() client.Response[T]) client.Response[T] {
	value, _ := catalogues.LoadOrStore(catalogueKey{cc: cc, name: name}, &catalogueEntry{})
	entry := value.(*catalogueEntry)

	entry.mu.Lock()
	defer entry.mu.Unlock()

	if cached, ok := entry.response.(client.Response[T]); ok {
		return cached
	}

	res := download()
	if !res.HasError() {
		entry.response = res
	}
	return res
}
