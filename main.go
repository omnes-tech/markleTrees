package main

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "log"
    "math/big"
    "net/http"

    // "context" and "math/big" are no longer strictly needed for the new Treap-based CMT,
    // but you can keep them if you're mixing with the old SimpleMerkleTree usage.
    "context"

    "merkleTrees/merkleGo"
)

type Response struct {
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, response Response) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(response)
}

func main() {
    // For your "simple" Merkle Tree (which uses go-merkletree-sql), you can keep this:
    hashFunc := func(data []byte) []byte {
        hash := sha256.Sum256(data)
        return hash[:]
    }

    // ---------------------
    // 1) Initialize Simple Merkle Tree (unchanged from your old code)
    // ---------------------
    simpleTree, err := merkleGo.NewSimpleMerkleTree(40, hashFunc)
    if err != nil {
        log.Fatalf("Failed to initialize Simple Merkle Tree: %v", err)
    }

    // ---------------------
    // 2) Initialize Treap-based Cartesian Merkle Tree
    // ---------------------
    // Instead of (depth, proofSize, hashFunc), we now just instantiate our treap-based CMT:
    cmt := merkleGo.NewCartesianMerkleTree()

    // ROUTES FOR Simple Merkle Tree (unchanged)
    http.HandleFunc("/simple/add", func(w http.ResponseWriter, r *http.Request) {
        key := big.NewInt(1)
        value := big.NewInt(100)

        err := simpleTree.Add(context.Background(), key, value)
        if err != nil {
            writeJSONResponse(w, http.StatusInternalServerError, Response{
                Message: "Failed to add to Simple Merkle Tree",
                Error:   err.Error(),
            })
            return
        }

        root, _ := simpleTree.GetRoot()
        writeJSONResponse(w, http.StatusOK, Response{
            Message: "Added to Simple Merkle Tree",
            Data: map[string]interface{}{
                "key":   key.String(),
                "value": value.String(),
                "root":  fmt.Sprintf("%x", root),
            },
        })
    })

    http.HandleFunc("/simple/proof", func(w http.ResponseWriter, r *http.Request) {
        key := big.NewInt(1)
        value := big.NewInt(100)

        proof, err := simpleTree.GenerateProof(context.Background(), key)
        if err != nil {
            writeJSONResponse(w, http.StatusInternalServerError, Response{
                Message: "Failed to generate proof for Simple Merkle Tree",
                Error:   err.Error(),
            })
            return
        }

        valid := simpleTree.VerifyProof(simpleTree.MerkleTree.Root(), proof, key, value)
        writeJSONResponse(w, http.StatusOK, Response{
            Message: "Generated proof for Simple Merkle Tree",
            Data: map[string]interface{}{
                "proof": proof,
                "valid": valid,
            },
        })
    })

    // ---------------------
    // ROUTES FOR Treap-based Cartesian Merkle Tree
    // ---------------------

    // /cmt/add: Insert a string "key" into our Treap
    http.HandleFunc("/cmt/add", func(w http.ResponseWriter, r *http.Request) {
        // For demonstration, let's add a fixed key, e.g. "hello"
        keyStr := "hello"

        err := cmt.Add([]byte(keyStr))
        if err != nil {
            writeJSONResponse(w, http.StatusInternalServerError, Response{
                Message: "Failed to add to Cartesian Merkle Tree",
                Error:   err.Error(),
            })
            return
        }

        root := cmt.GetRoot()
        writeJSONResponse(w, http.StatusOK, Response{
            Message: "Added to Cartesian Merkle Tree",
            Data: map[string]interface{}{
                "key":  keyStr,
                "root": hex.EncodeToString(root),
            },
        })
    })

    // /cmt/remove: Remove a given key from the Treap
    http.HandleFunc("/cmt/remove", func(w http.ResponseWriter, r *http.Request) {
        keyStr := "hello"

        err := cmt.Remove([]byte(keyStr))
        if err != nil {
            writeJSONResponse(w, http.StatusInternalServerError, Response{
                Message: "Failed to remove from Cartesian Merkle Tree",
                Error:   err.Error(),
            })
            return
        }

        root := cmt.GetRoot()
        writeJSONResponse(w, http.StatusOK, Response{
            Message: "Removed key from Cartesian Merkle Tree",
            Data: map[string]interface{}{
                "key":  keyStr,
                "root": hex.EncodeToString(root),
            },
        })
    })

    // /cmt/proof: Generate a proof for a given key, then verify it
    http.HandleFunc("/cmt/proof", func(w http.ResponseWriter, r *http.Request) {
        keyStr := "hello"

        // GenerateProof returns a struct with siblings, existence, etc.
        proof, err := cmt.GenerateProof([]byte(keyStr))
        if err != nil {
            writeJSONResponse(w, http.StatusInternalServerError, Response{
                Message: "Failed to generate proof for Cartesian Merkle Tree",
                Error:   err.Error(),
            })
            return
        }

        // Then we demonstrate local verification
        valid := cmt.VerifyProof([]byte(keyStr), proof)

        writeJSONResponse(w, http.StatusOK, Response{
            Message: "Generated proof for Cartesian Merkle Tree",
            Data: map[string]interface{}{
                "key":   keyStr,
                "proof": proof,
                "valid": valid,
            },
        })
    })

    // Start the HTTP server
    fmt.Println("Server running on port 8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
