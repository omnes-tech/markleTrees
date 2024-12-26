package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"

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
	hashFunc := func(data []byte) []byte {
		hash := sha256.Sum256(data)
		return hash[:]
	}

	// Initialize Simple Merkle Tree
	simpleTree, err := merkleGo.NewSimpleMerkleTree(40, hashFunc)
	if err != nil {
		log.Fatalf("Failed to initialize Simple Merkle Tree: %v", err)
	}

	// Initialize Cartesian Merkle Tree
	proofSize := 20
	cartesianTree, err := merkleGo.NewCartesianMerkleTree(40, proofSize, hashFunc)
	if err != nil {
		log.Fatalf("Failed to initialize Cartesian Merkle Tree: %v", err)
	}

	// Routes for Simple Merkle Tree
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

	// Routes for Cartesian Merkle Tree
	http.HandleFunc("/cmt/add", func(w http.ResponseWriter, r *http.Request) {
		key := big.NewInt(2)
		value := big.NewInt(200)

		err := cartesianTree.Add(context.Background(), key, value)
		if err != nil {
			writeJSONResponse(w, http.StatusInternalServerError, Response{
				Message: "Failed to add to Cartesian Merkle Tree",
				Error:   err.Error(),
			})
			return
		}

		root, _ := cartesianTree.GetRoot()
		writeJSONResponse(w, http.StatusOK, Response{
			Message: "Added to Cartesian Merkle Tree",
			Data: map[string]interface{}{
				"key":   key.String(),
				"value": value.String(),
				"root":  fmt.Sprintf("%x", root),
			},
		})
	})

	http.HandleFunc("/cmt/proof", func(w http.ResponseWriter, r *http.Request) {
		key := big.NewInt(2)
		value := big.NewInt(200)

		proof, err := cartesianTree.GenerateProof(context.Background(), key)
		if err != nil {
			writeJSONResponse(w, http.StatusInternalServerError, Response{
				Message: "Failed to generate proof for Cartesian Merkle Tree",
				Error:   err.Error(),
			})
			return
		}

		valid := cartesianTree.VerifyProof(cartesianTree.MerkleTree.Root(), proof, key, value)
		writeJSONResponse(w, http.StatusOK, Response{
			Message: "Generated proof for Cartesian Merkle Tree",
			Data: map[string]interface{}{
				"proof": proof,
				"valid": valid,
			},
		})
	})

	// Start the HTTP server
	fmt.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
