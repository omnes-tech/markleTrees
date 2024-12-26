package merkleGo

import (
    "crypto/sha256"
    "errors"
    "fmt"
    "bytes"
)

// TreapNode defines each node in the Cartesian Merkle Tree (Treap)
type TreapNode struct {
    Left       *TreapNode
    Right      *TreapNode
    Key        []byte
    Priority   []byte // Deterministic priority = keccak(key), or poseidon, etc.
    MerkleHash []byte
}

// CartesianMerkleTree holds the root of the Treap
type CartesianMerkleTree struct {
    Root *TreapNode
    // If you want to store desiredProofSize or a custom 3-arg hasher, do so here
    // e.g. HashFunc func(a, b, c []byte) []byte
}

// A minimal struct to demonstrate proof data
// In your on-chain code, you have something like
//   struct Proof { siblings[], siblingsLength, existence, key, nonExistenceKey }
type Proof struct {
    Existence bool
    Key       []byte
    Siblings  [][]byte
}

// 3-argument hasher using keccak256 (like _hash3 in Solidity)
func default3ArgHash(a, b, c []byte) []byte {
    // Ensure b < c by lexical comparison
    if bytes.Compare(b, c) > 0 {
        b, c = c, b
    }
    h := sha256.New()
    h.Write(a)
    h.Write(b)
    h.Write(c)
    return h.Sum(nil)
}

// Constructor
func NewCartesianMerkleTree() *CartesianMerkleTree {
    return &CartesianMerkleTree{}
}

// Insert a key into the Treap
func (cmt *CartesianMerkleTree) Add(key []byte) error {
    if len(key) == 0 {
        return errors.New("key cannot be empty")
    }
    priority := sha256.Sum256(key) // or a poseidon-based approach
    cmt.Root = cmt.insert(cmt.Root, key, priority[:])
    return nil
}

func (cmt *CartesianMerkleTree) insert(node *TreapNode, key, priority []byte) *TreapNode {
    if node == nil {
        newNode := &TreapNode{
            Key:        key,
            Priority:   priority,
        }
        // children = zero => hash(key, 0, 0)
        newNode.MerkleHash = default3ArgHash(key, make([]byte, 32), make([]byte, 32))
        return newNode
    }

    // BST property by key
    if bytes.Compare(key, node.Key) < 0 {
        node.Left = cmt.insert(node.Left, key, priority)
        // rotate if left child has bigger priority
        if bytes.Compare(node.Left.Priority, node.Priority) > 0 {
            node = cmt.rotateRight(node)
        }
    } else if bytes.Compare(key, node.Key) > 0 {
        node.Right = cmt.insert(node.Right, key, priority)
        // rotate if right child has bigger priority
        if bytes.Compare(node.Right.Priority, node.Priority) > 0 {
            node = cmt.rotateLeft(node)
        }
    } else {
        // key already exists => do nothing or update
        return node
    }

    node.MerkleHash = cmt.computeMerkleHash(node)
    return node
}

// Remove a key from the Treap
func (cmt *CartesianMerkleTree) Remove(key []byte) error {
    if len(key) == 0 {
        return errors.New("key cannot be empty")
    }
    // If the node doesn't exist, we'll do nothing or return error
    if cmt.Root == nil {
        return errors.New("tree is empty")
    }
    var removed bool
    cmt.Root, removed = cmt.remove(cmt.Root, key)
    if !removed {
        return fmt.Errorf("key %x not found", key)
    }
    return nil
}

func (cmt *CartesianMerkleTree) remove(node *TreapNode, key []byte) (*TreapNode, bool) {
    if node == nil {
        return nil, false
    }
    var removed bool
    cmp := bytes.Compare(key, node.Key)
    if cmp < 0 {
        node.Left, removed = cmt.remove(node.Left, key)
    } else if cmp > 0 {
        node.Right, removed = cmt.remove(node.Right, key)
    } else {
        // We found the node to remove
        removed = true
        // If no child or single child, just replace with non-nil child if any
        if node.Left == nil {
            return node.Right, true
        }
        if node.Right == nil {
            return node.Left, true
        }
        // If two children, we do rotation based on priority
        if bytes.Compare(node.Left.Priority, node.Right.Priority) < 0 {
            // rotateLeft
            node = cmt.rotateLeft(node)
            node.Left, _ = cmt.remove(node.Left, key)
        } else {
            // rotateRight
            node = cmt.rotateRight(node)
            node.Right, _ = cmt.remove(node.Right, key)
        }
    }

    if node != nil {
        node.MerkleHash = cmt.computeMerkleHash(node)
    }
    return node, removed
}

// Generate a proof for a given key (analogous to your Solidity library)
func (cmt *CartesianMerkleTree) GenerateProof(key []byte) (*Proof, error) {
    proof := &Proof{
        Existence: false,
        Key:       key,
        Siblings:  [][]byte{},
    }
    if cmt.Root == nil {
        // empty tree => can't exist
        return proof, nil
    }
    cmt.generateProofHelper(cmt.Root, key, proof)
    return proof, nil
}

// A DFS to find the key and collect siblings along the path
func (cmt *CartesianMerkleTree) generateProofHelper(node *TreapNode, key []byte, proof *Proof) {
    if node == nil {
        return
    }
    if bytes.Equal(node.Key, key) {
        // Found the node => push childLeftHash, childRightHash
        proof.Existence = true
        leftHash := make([]byte, 32)
        rightHash := make([]byte, 32)
        if node.Left != nil {
            leftHash = node.Left.MerkleHash
        }
        if node.Right != nil {
            rightHash = node.Right.MerkleHash
        }
        proof.Siblings = append(proof.Siblings, leftHash, rightHash)
        return
    }

    // If key < node.Key, go left
    if bytes.Compare(key, node.Key) < 0 {
        // We'll push (node.Key, rightChildHash) as siblings, for instance
        // This matches the pattern from your Solidity "someKey, otherChildHash"
        proof.Siblings = append(proof.Siblings, node.Key)

        rightHash := make([]byte, 32)
        if node.Right != nil {
            rightHash = node.Right.MerkleHash
        }
        proof.Siblings = append(proof.Siblings, rightHash)

        cmt.generateProofHelper(node.Left, key, proof)
    } else {
        // go right
        proof.Siblings = append(proof.Siblings, node.Key)

        leftHash := make([]byte, 32)
        if node.Left != nil {
            leftHash = node.Left.MerkleHash
        }
        proof.Siblings = append(proof.Siblings, leftHash)

        cmt.generateProofHelper(node.Right, key, proof)
    }
}

// VerifyProof: a simplistic local re-hash approach
func (cmt *CartesianMerkleTree) VerifyProof(key []byte, proof *Proof) bool {
    if !proof.Existence {
        // If the proof claims the key doesn't exist, then presumably it's false for membership
        return false
    }
    if len(key) == 0 || len(proof.Key) == 0 {
        return false
    }
    // We do a naive reconstruction approach similar to what you'd do in Solidity
    // For brevity, let's just rely on a function that matches your library's pattern
    computedRoot := cmt.rebuildFromProof(key, proof.Siblings)
    actualRoot := cmt.GetRoot()
    return bytes.Equal(computedRoot, actualRoot)
}

// This attempts to replicate the pattern in generateProofHelper
func (cmt *CartesianMerkleTree) rebuildFromProof(leaf []byte, siblings [][]byte) []byte {
    // Start from leaf
    current := leaf
    // If zero siblings => must be single-node tree => compare with hash(leaf, 0, 0)
    if len(siblings) == 2 && bytes.Equal(siblings[0], make([]byte, 32)) && bytes.Equal(siblings[1], make([]byte, 32)) {
        return default3ArgHash(leaf, siblings[0], siblings[1])
    }

    idx := 0
    for idx < len(siblings) {
        // We expect them in pairs => (nodeKey or childHash, theOtherChildHash)
        if idx+1 >= len(siblings) {
            break
        }
        s1 := siblings[idx]
        s2 := siblings[idx+1]
        idx += 2

        // Heuristics to see if s1 is the nodeKey or childHash
        // If s1 != 32 zero bytes and s1 != leaf => we treat it as nodeKey
        // Then decide if leaf < nodeKey => leaf is left, s2 is right
        // or leaf > nodeKey => s2 is left, leaf is right
        // If it doesn't look like a nodeKey, maybe it's the leftChildHash, rightChildHash
        // This is "guess-y" but follows the logic from generateProofHelper
        if len(s1) == 32 && !bytes.Equal(s1, make([]byte, 32)) && !bytes.Equal(s1, current) {
            // s1 is likely nodeKey
            if bytes.Compare(current, s1) < 0 {
                current = default3ArgHash(s1, current, s2)
            } else {
                current = default3ArgHash(s1, s2, current)
            }
        } else {
            // s1, s2 are leftHash, rightHash => nodeKey is current
            current = default3ArgHash(current, s1, s2)
        }
    }
    return current
}

// computeMerkleHash => hash(node.key, leftChildHash, rightChildHash)
func (cmt *CartesianMerkleTree) computeMerkleHash(node *TreapNode) []byte {
    var leftH, rightH []byte
    if node.Left != nil {
        leftH = node.Left.MerkleHash
    } else {
        leftH = make([]byte, 32)
    }
    if node.Right != nil {
        rightH = node.Right.MerkleHash
    } else {
        rightH = make([]byte, 32)
    }
    return default3ArgHash(node.Key, leftH, rightH)
}

// standard treap rotations
func (cmt *CartesianMerkleTree) rotateRight(y *TreapNode) *TreapNode {
    x := y.Left
    T2 := x.Right
    x.Right = y
    y.Left = T2

    y.MerkleHash = cmt.computeMerkleHash(y)
    x.MerkleHash = cmt.computeMerkleHash(x)
    return x
}

func (cmt *CartesianMerkleTree) rotateLeft(x *TreapNode) *TreapNode {
    y := x.Right
    T2 := y.Left
    y.Left = x
    x.Right = T2

    x.MerkleHash = cmt.computeMerkleHash(x)
    y.MerkleHash = cmt.computeMerkleHash(y)
    return y
}

// Return the root hash
func (cmt *CartesianMerkleTree) GetRoot() []byte {
    if cmt.Root == nil {
        return nil
    }
    return cmt.Root.MerkleHash
}
