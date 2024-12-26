# Merkle Trees in Go: Simple and Cartesian Implementations

This project demonstrates the implementation of two types of Merkle Trees using Golang:

1. **Simple Merkle Tree (SMT)**: A traditional Merkle Tree designed for cryptographic proofs of inclusion or exclusion, built on `github.com/iden3/go-merkletree-sql/v2`.
2. **Cartesian Merkle Tree (CMT)**: An advanced Merkle Tree combining features of **Binary Search Trees**, **Heaps**, and **Merkle Trees** (Treap-based). This is now **custom-implemented** without relying on `go-merkletree-sql` for the CMT logic, ensuring a fully deterministic, treap-like structure.

---

## Overview

Merkle Trees are cryptographic data structures widely used in blockchain systems, ZK rollups, and other applications requiring efficient and secure data verification. This project implements both SMT and CMT to explore their differences, benefits, and trade-offs.

- The **SMT** uses a standard binary Merkle approach via the `go-merkletree-sql` library.  
- The **CMT** is a **custom Treap-based** data structure that uses `priority = hash(key)` to maintain balance, storing data at **every node** rather than just the leaves.

---

## Features

### Simple Merkle Tree (SMT)
- Implements a traditional binary Merkle Tree structure using `go-merkletree-sql`.
- Stores data in leaf nodes only (library default).
- Efficient proof generation with `O(log(n))`.
- Ideal for simple membership proofs where you need a robust library-based approach.

### Cartesian Merkle Tree (CMT)
- **Treap-based** data structure combining:
  - **Binary Search Tree** ordering by `key`.
  - **Heap** property by `priority = sha256(key)`.
  - **Merkle Hash** (3-argument hash of `nodeKey`, `leftChildHash`, `rightChildHash`).
- Stores data in **every node**, requiring `n` total storage (instead of ~`2n` in a leaf-only design).
- Supports **insertion**, **removal**, and **proof generation/verification** with deterministic shape.
- Great for on-chain use cases or advanced ZK applications due to smaller on-chain footprint.

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

The server should start on port `8080`.

---

## API Endpoints

Below is a summary of the available routes in `main.go`. Some routes correspond to the **Simple Merkle Tree** (SMT) while others correspond to the **Cartesian Merkle Tree** (CMT).

### Simple Merkle Tree Routes (SMT)

1. **Add Data to SMT**
   - **Endpoint**: `POST /simple/add`
   - **Description**: Adds a leaf node to the Simple Merkle Tree (using the `go-merkletree-sql` library).
   - **Sample cURL**:
     ```bash
     curl -X POST http://localhost:8080/simple/add
     ```
   - **Sample Response**:
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
   - **Description**: Generates (and checks) a proof of inclusion for a key in the Simple Merkle Tree.
   - **Sample cURL**:
     ```bash
     curl -X GET http://localhost:8080/simple/proof
     ```
   - **Sample Response**:
     ```json
     {
       "message": "Generated proof for Simple Merkle Tree",
       "data": {
         "proof": "...",
         "valid": true
       }
     }
     ```

### Cartesian Merkle Tree Routes (CMT)

All the following routes operate on the **new Treap-based** CMT data structure implemented in `merkleGo/CartesianMerkleTree.go`.

1. **Add Data to CMT**
   - **Endpoint**: `POST /cmt/add`
   - **Description**: Inserts a **key** into the Treap-based Cartesian Merkle Tree.
   - **Sample cURL**:
     ```bash
     curl -X POST http://localhost:8080/cmt/add
     ```
   - **Sample Response**:
     ```json
     {
       "message": "Added to Cartesian Merkle Tree",
       "data": {
         "key": "hello",
         "root": "33f7b091695b0077db0d57f8981fc32d276d673f4a45b6fd70555027575e6458"
       }
     }
     ```

2. **Remove Data from CMT**
   - **Endpoint**: `POST /cmt/remove`
   - **Description**: Removes a **key** from the Treap-based CMT if it exists.
   - **Sample cURL**:
     ```bash
     curl -X POST http://localhost:8080/cmt/remove
     ```
   - **Sample Response**:
     ```json
     {
       "message": "Removed key from Cartesian Merkle Tree",
       "data": {
         "key": "hello",
         "root": "f9b34e3b0ba7e39273205e8039a1821dc44b5ab07fc606330327e7edf0fa96bf"
       }
     }
     ```

3. **Generate and Verify Proof in CMT**
   - **Endpoint**: `GET /cmt/proof`
   - **Description**: Generates a proof of inclusion for a given key in the Treap-based CMT, then verifies it locally.
   - **Sample cURL**:
     ```bash
     curl -X GET http://localhost:8080/cmt/proof
     ```
   - **Sample Response**:
     ```json
     {
       "message": "Generated proof for Cartesian Merkle Tree",
       "data": {
         "key": "hello",
         "proof": {
           "Existence": true,
           "Key": "aGVsbG8=", 
           "Siblings": ["..."]
         },
         "valid": true
       }
     }
     ```

---

## Comparison: SMT vs. CMT

| Feature              | Simple Merkle Tree (SMT)                                  | Cartesian Merkle Tree (CMT)                       |
|----------------------|-----------------------------------------------------------|---------------------------------------------------|
| **Storage**          | `2n` (leaf-only storage in typical binary Merkle trees)  | `n` (each node stores data)                       |
| **Node Structure**   | Data in leaves only                                      | Data in every node (BST + priority)               |
| **Proof Length**     | Typically fixed or library-dependent                     | Variable (prefix + suffix + possible rotations)   |
| **Complexity**       | `O(log(n))`                                              | `O(log(n))`, but shape is deterministic           |
| **Determinism**      | Standard Merkle BFS/DFS logic                            | Deterministic treap rotations via `hash(key)`     |
| **Use Cases**        | Simple inclusion/exclusion proofs                        | On-chain usage, advanced ZK proofs, efficient data updates |

---

## Code Details

### **Simple Merkle Tree (SMT)**
- Uses the `go-merkletree-sql` library to handle storage and proof generation.
- Leaf-based data storage.
- Great for quick integration when you need a stable library approach.

### **Cartesian Merkle Tree (CMT)**
- Implements a **Treap**: BST by `key`, heap by `priority = sha256(key)`.
- Each node maintains a **3-argument Merkle hash**: `hash(nodeKey, leftChildHash, rightChildHash)`.
- Supports **insertion**, **removal**, and **inclusion-proof generation** in purely custom Go code.
- A better fit if you need a **deterministic** data structure or want to mirror an **on-chain** Treap-based approach.

---

## Testing the Application

1. **Add a new leaf/node**  
   ```bash
   # Simple Merkle Tree:
   curl -X POST http://localhost:8080/simple/add

   # Cartesian Merkle Tree (Treap):
   curl -X POST http://localhost:8080/cmt/add
   ```

2. **Generate inclusion proof**  
   ```bash
   # Simple Merkle Tree:
   curl -X GET http://localhost:8080/simple/proof

   # Cartesian Merkle Tree:
   curl -X GET http://localhost:8080/cmt/proof
   ```

3. **Remove a node (CMT only)**  
   ```bash
   curl -X POST http://localhost:8080/cmt/remove
   ```
   Observe the updated root hash after successful removal.

4. Verify the generated proof in the **responses** (the server also does a local verification if you want to confirm correctness).

---

## Conclusion

This project highlights the key differences and use cases for **Simple** and **Treap-based Cartesian** Merkle Trees:

- **SMT** (Simple) is **straightforward** and ideal for basic applications or quick library usage.  
- **CMT** (Cartesian) offers **advanced** features: deterministic structure, optimized on-chain usage, and flexible proof logic—perfect for zero-knowledge or more complex cryptographic needs.

For more theoretical background, refer to the [Cartesian Merkle Tree: The New Breed](https://medium.com/@Arvolear/cartesian-merkle-tree-the-new-breed-a30b005ecf27).

---

## License

MIT License. See `LICENSE` for details.
```