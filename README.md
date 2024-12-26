# Merkle Trees in Go: Simple and Cartesian Implementations

This project demonstrates the implementation of two types of Merkle Trees using Golang and the `github.com/iden3/go-merkletree-sql/v2` library:
1. **Simple Merkle Tree (SMT)**: A traditional Merkle Tree designed for cryptographic proofs of inclusion or exclusion.
2. **Cartesian Merkle Tree (CMT)**: An advanced Merkle Tree combining features of Binary Search Trees, Heaps, and Merkle Trees to optimize storage and proof efficiency.

## Overview

Merkle Trees are cryptographic data structures widely used in blockchain systems, ZK rollups, and other applications requiring efficient and secure data verification. This project implements both SMT and CMT to explore their differences, benefits, and trade-offs.

---

## Features

### Simple Merkle Tree (SMT)
- Implements a traditional binary Merkle Tree structure.
- Stores data in leaf nodes only.
- Efficient proof generation with logarithmic complexity (`O(log(n))`).
- Suitable for applications where simplicity and fixed proof sizes are key.

### Cartesian Merkle Tree (CMT)
- A hybrid data structure combining:
  - **Binary Search Trees** for ordered key insertion.
  - **Heaps** for deterministic balancing using priority (`priority = hash(key)`).
  - **Merkle Trees** for cryptographic security.
- Stores data in every node, reducing storage size to `n` (compared to `2n` in SMT).
- Flexible proof sizes for advanced applications like ZK proofs.
- Ideal for on-chain operations where storage costs are high.

---

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/omnes-tech/merkleTrees.git
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Run the application:
   ```bash
   go run main.go
   ```

---

## API Endpoints

### Simple Merkle Tree Routes

1. **Add Data to SMT**
   - **Endpoint**: `POST /simple/add`
   - **Description**: Adds a leaf node to the Simple Merkle Tree.
   - **Response**:
     ```json
     {
       "message": "Added to Simple Merkle Tree",
       "data": {
         "key": "1",
         "value": "100",
         "root": "fbb94e2b0aa7e39273205e8039a1821dc44b5ab07fc606330327e7edf0fa96b0"
       }
     }
     ```

2. **Generate Proof in SMT**
   - **Endpoint**: `GET /simple/proof`
   - **Description**: Generates a proof of inclusion for a key.
   - **Response**:
     ```json
     {
       "message": "Generated proof for Simple Merkle Tree",
       "data": {
         "existence": true,
         "siblings": ["..."]
       }
     }
     ```

### Cartesian Merkle Tree Routes

1. **Add Data to CMT**
   - **Endpoint**: `POST /cmt/add`
   - **Description**: Adds a node to the Cartesian Merkle Tree.
   - **Response**:
     ```json
     {
       "message": "Added to Cartesian Merkle Tree",
       "data": {
         "node": "Cartesian Merkle Node",
         "root": "33f7b091695b0077db0d57f8981fc32d276d673f4a45b6fd70555027575e6458"
       }
     }
     ```

2. **Generate Proof in CMT**
   - **Endpoint**: `GET /cmt/proof`
   - **Description**: Generates a proof of inclusion for a key in the Cartesian Merkle Tree.
   - **Response**:
     ```json
     {
       "message": "Generated proof for Cartesian Merkle Tree",
       "data": {
         "existence": true,
         "siblings": ["..."]
       }
     }
     ```

---

## Comparison: SMT vs CMT

| Feature              | Simple Merkle Tree (SMT)                          | Cartesian Merkle Tree (CMT)                      |
|----------------------|---------------------------------------------------|--------------------------------------------------|
| **Storage**          | `2n`                                             | `n`                                              |
| **Node Structure**   | Data in leaves only                              | Data in every node                               |
| **Proof Length**     | Fixed                                            | Variable (prefix + suffix)                      |
| **Complexity**       | `O(log(n))`                                      | `O(log(n))`                                      |
| **Determinism**      | Deterministic for inclusion proofs               | Deterministic via `priority = hash(key)`        |
| **Use Cases**        | Simple inclusion/exclusion proofs                | ZK proofs, advanced cryptographic applications  |

---

## Code Details

### **Simple Merkle Tree (SMT)**
The SMT implementation focuses on:
1. **Leaf-Based Storage**: Only leaf nodes store data.
2. **Merkle Proofs**: Inclusion proofs are calculated by traversing the tree from root to leaf.
3. **Hash Function**: `sha256` is used for hashing.

### **Cartesian Merkle Tree (CMT)**
The CMT implementation introduces:
1. **Deterministic Balancing**: Nodes are balanced using `priority = hash(key)`.
2. **Compact Storage**: Stores values in every node, reducing overall storage.
3. **Proof Composition**: Proofs include both prefix (path) and suffix (hash of siblings).

---

## Testing the Application

1. Add a new leaf or node:
   ```bash
   curl -X POST http://localhost:8080/simple/add
   curl -X POST http://localhost:8080/cmt/add
   ```

2. Generate inclusion proof:
   ```bash
   curl -X GET http://localhost:8080/simple/proof
   curl -X GET http://localhost:8080/cmt/proof
   ```

3. Verify the generated proof in the logs or responses to ensure correctness.

---

## Conclusion

This project highlights the key differences and use cases for Simple and Cartesian Merkle Trees:
- **SMT** is straightforward and ideal for basic applications.
- **CMT** offers advanced features, optimized storage, and flexibility for ZK proofs.

For more theoretical background, refer to the [Cartesian Merkle Tree: The New Breed](https://medium.com/@Arvolear/cartesian-merkle-tree-the-new-breed-a30b005ecf27).

---

## License

MIT License. See `LICENSE` for details.