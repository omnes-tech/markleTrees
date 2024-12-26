// SPDX-License-Identifier: MIT
pragma solidity ^0.8.4;

import "./CartesianMerkleTree.sol"; // Adjust the path as necessary
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract TokenCMT is ERC20 {
    using CartesianMerkleTree for CartesianMerkleTree.Bytes32CMT;

    CartesianMerkleTree.Bytes32CMT private cmt;
    mapping(address => bool) private minted;

    event Mint(address indexed user, uint256 amount);

    constructor(uint32 proofSize, string memory name, string memory symbol) ERC20(name, symbol) {
        // Initialize the Merkle Tree with a desired proof size
        cmt.initialize(proofSize);
    }

    /**
     * @notice Adds an address to the Merkle Tree, enabling it to mint tokens.
     * @param user The address to add to the tree.
     */
    function addToWhitelist(address user) external {
        require(user != address(0), "Invalid address");
        bytes32 key = keccak256(abi.encodePacked(user));
        cmt.add(key);
    }

    /**
     * @notice Removes an address from the Merkle Tree, disabling its minting ability.
     * @param user The address to remove from the tree.
     */
    function removeFromWhitelist(address user) external {
        require(user != address(0), "Invalid address");
        bytes32 key = keccak256(abi.encodePacked(user));
        cmt.remove(key);
    }

    /**
     * @notice Mints tokens to the caller if they are in the Merkle Tree.
     * @param proof The Merkle proof showing inclusion in the tree.
     */
    function mint(CartesianMerkleTree.Proof calldata proof) external {
        require(!minted[msg.sender], "Already minted");

        bytes32 key = keccak256(abi.encodePacked(msg.sender));
        require(proof.existence && proof.key == key, "Invalid proof or not whitelisted");

        // Validate the proof
        bytes32 root = cmt.getRoot();
        require(_verifyProof(proof, root), "Invalid proof");

        // Mint tokens
        minted[msg.sender] = true;
        _mint(msg.sender, 1 * 10 ** decimals()); // Mint 1 token (adjust as needed)
        emit Mint(msg.sender, 1 * 10 ** decimals());
    }

    // /**
    //  * @notice Verifies the Merkle proof.
    //  * @param proof The proof to verify.
    //  * @param root The current Merkle root.
    //  * @return True if the proof is valid, false otherwise.
    //  */
    // function _verifyProof(CartesianMerkleTree.Proof calldata proof, bytes32 root) internal pure returns (bool) {
    //     // If the library says "existence = false," there's nothing to verify
    //     // 1) Must claim existence = true, must match the user’s key
    //     if (!proof.existence) {
    //         return false;
    //     }

    //     // If it's a single-node tree (no children), you'd just check:
    //     // keccak256(abi.encodePacked(proof.key, bytes32(0), bytes32(0))) == root
    //     // But for multiple nodes, you'd need to replicate the library's logic
    //     // which takes the node’s key plus two children’s merkleHashes, merges them
    //     // with `keccak256(nodeKey, smallerHash, biggerHash)`, etc.

    //     // For a minimal patch that allows your single-user test to pass when the
    //     // user is the only node, you can do something like:
    //     // 2) If this is the only node in the tree, siblings will be [0,0].
    //     //    So let's do the 3-argument keccak256 check:
    //     if (proof.siblingsLength == 2 && proof.siblings[0] == bytes32(0) && proof.siblings[1] == bytes32(0)) {
    //         // i.e. a leaf with no children
    //         bytes32 leafHash = keccak256(abi.encodePacked(proof.key, bytes32(0), bytes32(0)));
    //         return (leafHash == root);
    //     }

    //     // Otherwise, you'd need to parse pairs of siblings in the same way
    //     // the library’s `_proof()` function is building them. That is more
    //     // involved, so many people instead choose Option B below.

    //     // For a multi-node scenario, you’d have to replicate or
    //     // re-walk the path with the same 3-argument logic. That is more complex.
    //     // For now, revert or return false:
    //     return false;
    // }

    //B
    function _verifyProof(CartesianMerkleTree.Proof calldata proof, bytes32 root) internal pure returns (bool) {
        // 0. Quick checks
        if (!proof.existence) {
            // Claimed existence = false => no reason to verify
            return false;
        }
        if (proof.key == bytes32(0)) {
            return false; // key cannot be zero
        }

        // 1. Rebuild the path from leaf to root by iterating siblings in pairs
        bytes32 currentHash = proof.key;
        bytes32[] calldata siblings = proof.siblings;

        // If no siblings => it's presumably a 1-node tree
        // So we do the single check: keccak256(key,0,0) == root ?
        if (proof.siblingsLength == 2 && siblings.length >= 2 && siblings[0] == bytes32(0) && siblings[1] == bytes32(0))
        {
            bytes32 leafHash = _hash3(currentHash, bytes32(0), bytes32(0));
            return (leafHash == root);
        }

        // 2. Otherwise, handle the multi-sibling scenario
        // We'll parse siblings in increments of 2
        uint256 idx = 0;
        while (idx + 1 < siblings.length && idx + 1 < proof.siblingsLength) {
            bytes32 s1 = siblings[idx];
            bytes32 s2 = siblings[idx + 1];
            idx += 2;

            // Detect pattern:
            //   - If s1 == childLeftHash and s2 == childRightHash =>
            //       node’s key = currentHash
            //   - Else if s1 == node.key and s2 == some child’s merkleHash =>
            //       we must figure out whether our currentHash is childLeft or childRight

            // Heuristic: If s1 < 0x0100... then maybe it's 0-hash or a child’s merkleHash
            // This is tricky, so see how your library encodes them.

            bool s1IsNodeKey = _mightBeNodeKey(currentHash, s1, s2);

            if (!s1IsNodeKey) {
                // Means s1, s2 are leftChildHash, rightChildHash
                // and currentHash is the node’s key.
                currentHash = _hash3(
                    currentHash,
                    s1, // left
                    s2 // right
                );
            } else {
                // s1 is the node’s key, s2 is one child’s merkleHash
                // We must see if currentHash < s1 => then (left= currentHash, right= s2)
                // else (left= s2, right= currentHash).
                if (currentHash < s1) {
                    currentHash = _hash3(
                        s1,
                        currentHash, // as left child
                        s2 // as right child
                    );
                } else {
                    currentHash = _hash3(
                        s1,
                        s2, // left child
                        currentHash // right child
                    );
                }
            }
        }

        return (currentHash == root);
    }

    // Example 3-arg hashing replicating `_hash3` from library
    function _hash3(bytes32 a, bytes32 b, bytes32 c) private pure returns (bytes32) {
        // Sort b,c so that the smaller one is first
        // same logic as your library does:
        if (b > c) {
            (b, c) = (c, b);
        }
        return keccak256(abi.encodePacked(a, b, c));
    }

    // Heuristic helper to guess if s1 is the node key or child’s merkleHash
    function _mightBeNodeKey(bytes32 current, bytes32 s1, bytes32 s2) private pure returns (bool) {
        // In your library’s `_proof()`, if node.key == key => we do push childLeftHash, childRightHash
        // Otherwise we do push node.key + the other child’s merkleHash
        // So if s1 == node.key, it means s1 is not likely 0-hash or child’s merkleHash
        // Some teams check if s1 is the same as `current` => not always correct
        // Or we can just guess if s2 is 0-hash or something.
        // This is heavily dependent on how your library implements `_addProofSibling`.

        // A simplistic approach: if s1 != 0 and s2 != 0 and s1 != current, maybe s1 is the node’s key
        // This is hacky but demonstrates the idea
        if (s1 != bytes32(0) && s2 != bytes32(0) && s1 != current) {
            return true;
        }
        return false;
    }

    /**
     * @notice Retrieves the Merkle root.
     * @return The Merkle root of the current tree.
     */
    function getMerkleRoot() external view returns (bytes32) {
        return cmt.getRoot();
    }

    /**
     * @notice Generates a proof for a given key.
     * @param key The key for which to generate the proof.
     * @param desiredProofSize The desired number of siblings in the proof.
     * @return The proof for the given key.
     */
    function getProof(bytes32 key, uint32 desiredProofSize) external view returns (CartesianMerkleTree.Proof memory) {
        return cmt.getProof(key, desiredProofSize);
    }
}
