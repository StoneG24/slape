/*
Package vectorstore is used to create independant context information for each disscussion.
Vectorstores allow us to index information and keep relevant bits for generation.
*/
package vectorstore

import ()

// Utilize HNSW algorithim to create a vector for Internet searching
type VectoreStore struct {
}

// This is a simple function to get us going and will need more work in future to implement
// correctly. Apparently this has some openai compat
func (v *VectoreStore) Startup() {
}
