// SPDX-License-Identifier: MIT
pragma solidity ^0.8.27;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/utils/cryptography/MerkleProof.sol";

contract TokenMT is ERC20 {
    using MerkleProof for bytes32[];

    bytes32 public immutable merkleRoot;
    mapping(address => bool) public hasMinted;

    event Mint(address indexed user, uint256 amount);

    constructor(string memory name, string memory symbol, bytes32 _merkleRoot) ERC20(name, symbol) {
        require(_merkleRoot != bytes32(0), "Merkle root cannot be zero");
        merkleRoot = _merkleRoot;
    }

    /**
     * @notice Mints tokens if the sender provides a valid proof of inclusion in the Merkle Tree.
     * @param amount The amount of tokens to mint.
     * @param proof The Merkle proof showing inclusion in the tree.
     */
    function mint(uint256 amount, bytes32[] calldata proof) external {
        require(!hasMinted[msg.sender], "Already minted");
        bytes32 leaf = keccak256(abi.encodePacked(msg.sender, amount));
        require(proof.verify(merkleRoot, leaf), "Invalid proof");

        hasMinted[msg.sender] = true;
        _mint(msg.sender, amount);
        emit Mint(msg.sender, amount);
    }
}
