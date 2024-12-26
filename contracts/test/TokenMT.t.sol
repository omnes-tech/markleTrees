// SPDX-License-Identifier: MIT
pragma solidity ^0.8.4;

import "forge-std/Test.sol";
import "src/TokenMT.sol";
import "@openzeppelin/contracts/utils/cryptography/MerkleProof.sol";

// Helper library exactly matching OpenZeppelin's "commutativeKeccak256"
library LocalHashes {
    function commutativeKeccak256(bytes32 a, bytes32 b) internal pure returns (bytes32) {
        // Sort the pair
        return a < b ? keccak256(abi.encode(a, b)) : keccak256(abi.encode(b, a));
    }
}

contract TokenMTTest is Test {
    TokenMT public token;

    // We'll make a little 4-leaf Merkle tree:
    // leaf0, leaf1, leaf2, leaf3 => we combine them in pairs => then combine again => root
    // We'll store one of the leaves/proofs for user=0x123, amount=100 ether
    bytes32 public leafUser;
    bytes32[] public validProofForUser;
    uint256 public constant userAmount = 100 ether;
    address public constant userAddr = address(0x123);

    // We'll also store the final merkle root here
    bytes32 public root;

    function setUp() public {
        /**
         * 1) Define 4 leaves:
         *    leaf0 = keccak256( (0x123), (100 ether) )
         *    leaf1 = keccak256( arbitrary )
         *    leaf2 = keccak256( arbitrary )
         *    leaf3 = keccak256( arbitrary )
         */
        leafUser = keccak256(abi.encodePacked(userAddr, userAmount));
        bytes32 leaf1 = keccak256(abi.encodePacked(address(0xABC), uint256(200 ether)));
        bytes32 leaf2 = keccak256(abi.encodePacked(address(0xDEF), uint256(300 ether)));
        bytes32 leaf3 = keccak256(abi.encodePacked(address(0x999), uint256(400 ether)));

        // 2) Pair them up in level 1:
        //    node0 = H( leafUser, leaf1 )
        //    node1 = H( leaf2, leaf3 )
        bytes32 node0 = LocalHashes.commutativeKeccak256(leafUser, leaf1);
        bytes32 node1 = LocalHashes.commutativeKeccak256(leaf2, leaf3);

        // 3) Pair them up in level 2 => the root:
        //    root = H(node0, node1)
        root = LocalHashes.commutativeKeccak256(node0, node1);

        // 4) The proof for leafUser is [leaf1, node1].
        //    Explanation:
        //      - First sibling at the bottom is leaf1.
        //      - Second sibling at the top level is node1.
        validProofForUser = new bytes32[](2);
        validProofForUser[0] = leaf1;
        validProofForUser[1] = node1;

        // 5) Deploy TokenMT with the new real merkle root
        token = new TokenMT("TestToken", "TTK", root);
    }

    function testMintValidProof() public {
        // We'll use the user=0x123 from setUp and the validProofForUser
        vm.prank(userAddr);
        token.mint(userAmount, validProofForUser);

        assertEq(token.balanceOf(userAddr), userAmount);
        assertTrue(token.hasMinted(userAddr));
    }

    function testMintInvalidProof() public {
        // Some random user + random amounts
        address user = address(0x123);
        uint256 amount = 100 ether;

        // Invalid proof
        bytes32[] memory badProof = new bytes32[](2);
        badProof[0] = bytes32(0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef);
        badProof[1] = bytes32(0xfeedbeeffeedbeeffeedbeeffeedbeeffeedbeeffeedbeeffeedbeeffeedbeef);

        // Prank as the user
        vm.prank(user);

        vm.expectRevert("Invalid proof");
        token.mint(amount, badProof);
    }

    function testDoubleMint() public {
        // First mint with valid proof
        vm.prank(userAddr);
        token.mint(userAmount, validProofForUser);

        // Attempt to mint again from same user
        vm.prank(userAddr);
        vm.expectRevert("Already minted");
        token.mint(userAmount, validProofForUser);
    }
}
