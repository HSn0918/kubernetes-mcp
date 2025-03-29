package fake

import (
	"context"

	"github.com/hsn0918/kubernetes-mcp/pkg/client"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime" // Standard way to add schemes
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// FakeKubernetesClient is a mock client for testing.
// It embeds ctrlclient.Client to avoid manually implementing all methods.
type FakeKubernetesClient struct {
	ctrlclient.Client // Embed the fake client directly
}

// Ensure FakeKubernetesClient implements KubernetesClient interface
// (We redefine the interface methods to ensure they are present,
// although embedding covers most implementations).
var _ client.KubernetesClient = &FakeKubernetesClient{}

// NewFakeClient creates a new fake Kubernetes client with initialized scheme.
func NewFakeClient(initObjs ...ctrlclient.Object) *FakeKubernetesClient {
	scheme := runtime.NewScheme()

	// Register standard Kubernetes types (IMPORTANT!)
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	// Add other schemes if needed (e.g., custom resources, specific API groups)
	// utilruntime.Must(corev1.AddToScheme(scheme))
	// utilruntime.Must(appsv1.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(initObjs...). // For Get/List tests
		// WithStatusSubresource(initObjs...) // If testing status updates
		Build()

	return &FakeKubernetesClient{
		Client: fakeClient,
	}
}

// The embedded fake client already implements these methods.
// We list them here explicitly primarily for interface satisfaction clarity
// if we weren't embedding the client interface directly in KubernetesClient.
// Since KubernetesClient currently *is* client.Client, embedding is sufficient.

// Create forwards to the embedded fake client.
func (f *FakeKubernetesClient) Create(ctx context.Context, obj ctrlclient.Object, opts ...ctrlclient.CreateOption) error {
	return f.Client.Create(ctx, obj, opts...)
}

// Delete forwards to the embedded fake client.
func (f *FakeKubernetesClient) Delete(ctx context.Context, obj ctrlclient.Object, opts ...ctrlclient.DeleteOption) error {
	return f.Client.Delete(ctx, obj, opts...)
}

// Update forwards to the embedded fake client.
func (f *FakeKubernetesClient) Update(ctx context.Context, obj ctrlclient.Object, opts ...ctrlclient.UpdateOption) error {
	return f.Client.Update(ctx, obj, opts...)
}

// Patch forwards to the embedded fake client.
func (f *FakeKubernetesClient) Patch(ctx context.Context, obj ctrlclient.Object, patch ctrlclient.Patch, opts ...ctrlclient.PatchOption) error {
	return f.Client.Patch(ctx, obj, patch, opts...)
}

// Get forwards to the embedded fake client.
func (f *FakeKubernetesClient) Get(ctx context.Context, key ctrlclient.ObjectKey, obj ctrlclient.Object, opts ...ctrlclient.GetOption) error {
	// Note: fake client GetOptions were added later. Ensure your controller-runtime version supports them if used.
	// The base Get method without GetOption exists in older versions.
	// return f.Client.Get(ctx, key, obj, opts...) // Use this if GetOption is supported
	return f.Client.Get(ctx, key, obj) // Base Get method
}

// List forwards to the embedded fake client.
func (f *FakeKubernetesClient) List(ctx context.Context, list ctrlclient.ObjectList, opts ...ctrlclient.ListOption) error {
	return f.Client.List(ctx, list, opts...)
}

// DeleteAllOf forwards to the embedded fake client.
func (f *FakeKubernetesClient) DeleteAllOf(ctx context.Context, obj ctrlclient.Object, opts ...ctrlclient.DeleteAllOfOption) error {
	return f.Client.DeleteAllOf(ctx, obj, opts...)
}

// Status forwards to the embedded fake client's Status writer.
func (f *FakeKubernetesClient) Status() ctrlclient.StatusWriter {
	return f.Client.Status()
}

// Scheme returns the scheme used by the fake client.
func (f *FakeKubernetesClient) Scheme() *runtime.Scheme {
	return f.Client.Scheme()
}

// RESTMapper returns the RESTMapper used by the fake client.
func (f *FakeKubernetesClient) RESTMapper() meta.RESTMapper {
	return f.Client.RESTMapper()
}

// SubResource forwards to the embedded fake client.
func (f *FakeKubernetesClient) SubResource(subResource string) ctrlclient.SubResourceClient {
	return f.Client.SubResource(subResource)
}

// GroupVersionKindFor forwards to the embedded fake client.
func (f *FakeKubernetesClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return f.Client.GroupVersionKindFor(obj)
}

// IsObjectNamespaced forwards to the embedded fake client.
func (f *FakeKubernetesClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return f.Client.IsObjectNamespaced(obj)
}

// ---- Helper methods for testing specific scenarios ----

// SimulateNotFoundOnGet configures the client to return NotFound for a specific Get call.
// This requires more advanced mocking capabilities, often provided by libraries like mockery
// or by customizing the fake client's reactor chain, which is complex.
// The standard fake client doesn't easily support conditional errors per call without setup.
// func (f *FakeKubernetesClient) SimulateNotFoundOnGet(key ctrlclient.ObjectKey, gvk schema.GroupVersionKind) {
// 	// Implementation would involve adding a reactor to the fake client's tracker.
// 	// reactor := &testingfake.ReactionFunc(func(action testingfake.Action) (handled bool, ret runtime.Object, err error) {
// 	// 	if getAction, ok := action.(testingfake.GetAction); ok {
// 	// 		if getAction.GetNamespace() == key.Namespace && getAction.GetName() == key.Name && getAction.GetResource() == gvk.GroupVersion().WithResource(strings.ToLower(gvk.Kind)+"s") {
// 	// 			return true, nil, kerrors.NewNotFound(gvk.GroupResource(), key.Name)
// 	// 		}
// 	// 	}
// 	// 	return false, nil, nil
// 	// })
//   // Requires access to the underlying fake clientset if using client-go's fake directly
// 	// f.FakeClient.PrependReactor("get", "*", reactor)
// }
