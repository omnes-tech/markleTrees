package merkleGo

import (
	"context"
	"crypto/sha256"
	"errors"
	"math/big"

	"github.com/iden3/go-merkletree-sql/v2"
	"github.com/iden3/go-merkletree-sql/v2/db/memory"
)

type SimpleMerkleTree struct {
	MerkleTree *merkletree.MerkleTree
	HashFunc   func(data []byte) []byte
	Leaves     [][]byte // To store the raw leaves
}

// Initialize the Simple Merkle Tree
func NewSimpleMerkleTree(depth int, hashFunc func(data []byte) []byte) (*SimpleMerkleTree, error) {
	ctx := context.Background()
	storage := memory.NewMemoryStorage()
	tree, err := merkletree.NewMerkleTree(ctx, storage, depth)
	if err != nil {
		return nil, err
	}
	return &SimpleMerkleTree{MerkleTree: tree, HashFunc: hashFunc}, nil
}

// Add a leaf to the tree
func (smt *SimpleMerkleTree) AddLeaf(data []byte) {
	hash := sha256.Sum256(data)
	smt.Leaves = append(smt.Leaves, hash[:])
}

// Add a key-value pair to the tree
func (smt *SimpleMerkleTree) Add(ctx context.Context, key *big.Int, value *big.Int) error {
	return smt.MerkleTree.Add(ctx, key, value)
}

// Generate a proof for a key
func (smt *SimpleMerkleTree) GenerateProof(ctx context.Context, key *big.Int) (*merkletree.Proof, error) {
	root := smt.MerkleTree.Root()
	proof, _, err := smt.MerkleTree.GenerateProof(ctx, key, root)
	if err != nil {
		return nil, err
	}
	return proof, nil
}

// Verify a proof
func (smt *SimpleMerkleTree) VerifyProof(root *merkletree.Hash, proof *merkletree.Proof, key *big.Int, value *big.Int) bool {
	return merkletree.VerifyProof(root, proof, key, value)
}

// GetRoot retrieves the current Merkle root from the tree
func (smt *SimpleMerkleTree) GetRoot() ([]byte, error) {
	root := smt.MerkleTree.Root()
	if root == nil {
		return nil, errors.New("Merkle tree root is nil")
	}
	return root.BigInt().Bytes(), nil // Convert *big.Int to bytes
}
