package merkleGo

import (
	"context"
	"crypto/sha256"
	"errors"
	"math/big"

	"github.com/iden3/go-merkletree-sql/v2"
	"github.com/iden3/go-merkletree-sql/v2/db/memory"
)

type CartesianMerkleTree struct {
	MerkleTree *merkletree.MerkleTree
	HashFunc   func(data []byte) []byte
	Nodes      [][]byte // To store the raw nodes
	MaxElements int      // Limit based on proof size
}

// Initialize the Cartesian Merkle Tree
func NewCartesianMerkleTree(depth int, proofSize int, hashFunc func(data []byte) []byte) (*CartesianMerkleTree, error) {
	ctx := context.Background()
	storage := memory.NewMemoryStorage()
	tree, err := merkletree.NewMerkleTree(ctx, storage, depth)
	if err != nil {
		return nil, err
	}
	return &CartesianMerkleTree{MerkleTree: tree, HashFunc: hashFunc, MaxElements: 1 << proofSize}, nil
}

// Add a node to the tree
func (cmt *CartesianMerkleTree) AddNode(data []byte) error {
	if len(cmt.Nodes) >= cmt.MaxElements {
		return errors.New("tree is full")
	}
	hash := sha256.Sum256(data)
	cmt.Nodes = append(cmt.Nodes, hash[:])
	return nil
}

// Add a key-value pair to the tree
func (cmt *CartesianMerkleTree) Add(ctx context.Context, key *big.Int, value *big.Int) error {
	return cmt.MerkleTree.Add(ctx, key, value)
}

// Generate a proof for a key
func (cmt *CartesianMerkleTree) GenerateProof(ctx context.Context, key *big.Int) (*merkletree.Proof, error) {
	root := cmt.MerkleTree.Root()
	proof, _, err := cmt.MerkleTree.GenerateProof(ctx, key, root)
	if err != nil {
		return nil, err
	}
	return proof, nil
}

// Verify a proof
func (cmt *CartesianMerkleTree) VerifyProof(root *merkletree.Hash, proof *merkletree.Proof, key *big.Int, value *big.Int) bool {
	return merkletree.VerifyProof(root, proof, key, value)
}

// GetRoot retrieves the current Merkle root from the tree
func (cmt *CartesianMerkleTree) GetRoot() ([]byte, error) {
	root := cmt.MerkleTree.Root()
	if root == nil {
		return nil, errors.New("Merkle tree root is nil")
	}
	return root.BigInt().Bytes(), nil // Convert *big.Int to bytes
}
