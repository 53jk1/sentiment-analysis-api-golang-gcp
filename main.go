package main

import (
	language "cloud.google.com/go/language/apiv1"
	"context"
	"encoding/json"
	"log"
	"net/http"

	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
)

type SentimentRequest struct {
	Text string `json:"text"`
}

type SentimentResponse struct {
	Sentiment      string  `json:"sentiment"`
	SentimentScore float32 `json:"sentiment_score"`
}

func main() {
	http.HandleFunc("/analyze", analyzeHandler)
	http.HandleFunc("/healthcheck", healthcheckHandler)
	http.HandleFunc("/docs", docsHandler)

	log.Println("Starting Sentiment Analysis API server on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var req SentimentRequest
	if err := decoder.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	client, err := language.NewClient(ctx)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := client.AnalyzeSentiment(ctx, &languagepb.AnalyzeSentimentRequest{
		Document: &languagepb.Document{
			Source: &languagepb.Document_Content{
				Content: req.Text,
			},
			Type:     languagepb.Document_PLAIN_TEXT,
			Language: "en",
		},
	})
	if err != nil {
		log.Printf("Failed to analyze sentiment: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var sentiment string
	var sentimentScore float32
	if sc := resp.DocumentSentiment.Score; sc > 0 {
		sentiment = "positive"
		sentimentScore = sc
	} else if sc < 0 {
		sentiment = "negative"
		sentimentScore = -sc
	} else {
		sentiment = "neutral"
	}

	json.NewEncoder(w).Encode(SentimentResponse{
		Sentiment:      sentiment,
		SentimentScore: sentimentScore,
	})
}

func healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func docsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	swagger := `{
	"swagger": "2.0",
	"info": {
		"title": "Sentiment Analysis API",
		"description": "A simple API to analyze the sentiment of a text",
		"version": "1.0.0"
	},
	"host": "localhost:8080",
	"basePath": "/",
	"paths": {
		"/analyze": {
			"post": {
				"summary": "Analyze the sentiment of a text",
				"description": "Analyze the sentiment of a text",
				"consumes": [
					"application/json"
				],
				"produces": [
					"application/json"
				],	
				"parameters": [	
					{
						"name": "body",
						"in": "body",	
						"schema": {
							"$ref": "#/definitions/SentimentRequest"
						}	
					}
				],	
				"responses": {
					"200": {
						"description": "Success",
						"schema": {
							"$ref": "#/definitions/SentimentResponse"	
						}	
					},
					"400": {
						"description": "Bad Request"	
					},	
					"405": {
						"description": "Method Not Allowed"
					}
				}	
			}
		},	
		"/healthcheck": {	
			"get": {	
				"summary": "Healthcheck",	
				"description": "Healthcheck",	
				"produces": [	
					"application/json"
				],	
				"responses": {	
					"200": {
						"description": "Success"
					},
					"405": {
						"description": "Method Not Allowed"	
					}
				}	
			}	
		}	
	},	
	"definitions": {	
		"SentimentRequest": {	
			"type": "object",
			"properties": {	
				"text": {	
					"type": "string"	
				}	
			}	
		},	
		"SentimentResponse": {	
			"type": "object",	
			"properties": {	
				"sentiment": {	
					"type": "string"
				},	
				"sentiment_score": {	
					"type": "number"	
				}	
			}	
		}	
	}
}`

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(swagger))
}
