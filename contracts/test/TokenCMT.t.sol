// SPDX-License-Identifier: MIT
pragma solidity ^0.8.4;

import "forge-std/Test.sol";
import "src/TokenCMT.sol";
import "src/CartesianMerkleTree.sol";

contract TokenCMTtest is Test {
    TokenCMT public merkleMint;
    uint32 public proofSize = 10;

    function setUp() public {
        // Initialize the MerkleMint contract
        merkleMint = new TokenCMT(proofSize, "TestToken", "TTK");
    }

    function testAddToWhitelist() public {
        address user = address(0x123);
        merkleMint.addToWhitelist(user);
        bytes32 root = merkleMint.getMerkleRoot();
        assert(root != bytes32(0)); // Ensure root is updated
    }

    function testMintWithProof() public {
        address user = address(0x123);

        // Add user to the whitelist
        merkleMint.addToWhitelist(user);

        // Fetch updated Merkle root
        bytes32 root = merkleMint.getMerkleRoot();
        assert(root != bytes32(0));

        // Generate proof from the updated tree
        bytes32 key = keccak256(abi.encodePacked(user));
        CartesianMerkleTree.Proof memory proof = merkleMint.getProof(key, proofSize);

        // Validate proof parameters
        assertEq(proof.root, root, "Proof root does not match Merkle root");
        assertEq(proof.key, key, "Proof key does not match user key");

        // Perform mint operation
        vm.prank(user);
        merkleMint.mint(proof);
        assertEq(merkleMint.balanceOf(user), 1 * 10 ** 18); // Verify token balance
    }

    function testMintFailWithoutProof() public {
        address user = address(0x123);

        // Attempt minting without a valid proof
        CartesianMerkleTree.Proof memory proof;
        vm.prank(user);
        vm.expectRevert("Invalid proof or not whitelisted");
        merkleMint.mint(proof);
    }
}
