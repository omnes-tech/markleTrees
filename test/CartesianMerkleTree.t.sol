// SPDX-License-Identifier: MIT
pragma solidity ^0.8.4;

import "forge-std/Test.sol";
import "../src/CartesianMerkleTree.sol"; // Ensure the path is correct.

contract CartesianMerkleTreeTest is Test {
    using CartesianMerkleTree for CartesianMerkleTree.UintCMT;

    CartesianMerkleTree.UintCMT internal uintTreaple;

    function setUp() public {
        // Increase the proof size during initialization
        uintTreaple.initialize(10); // Adjust the proof size as needed
    }

    function testAddAndRemoveNodes() public {
        uintTreaple.add(10);
        uintTreaple.add(20);
        uintTreaple.add(15);

        bytes32 root = uintTreaple.getRoot();
        assert(root != bytes32(0));

        CartesianMerkleTree.Proof memory proof = uintTreaple.getProof(15, 10);
        assert(proof.existence == true);
        assert(proof.key == bytes32(uint256(15)));

        uintTreaple.remove(15);
        bytes32 newRoot = uintTreaple.getRoot();
        assert(newRoot != root);
    }

    function testProofNonExistence() public {
        uintTreaple.add(5);
        uintTreaple.add(25);

        CartesianMerkleTree.Proof memory proof = uintTreaple.getProof(15, 10);
        assert(proof.existence == false);
    }

    function testCustomHasher() public {
        // Set the custom hash function
        uintTreaple.setHasher(customHash);

        // Add elements and validate functionality
        uintTreaple.add(100);
        assert(uintTreaple.isCustomHasherSet());
    }

    // Custom hash function matching the required signature
    function customHash(bytes32 a, bytes32 b, bytes32 c) public pure returns (bytes32) {
        return keccak256(abi.encodePacked(a, b, c));
    }
}
