package utils

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// string -> objectID
func StringToObjectID(idStr string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(idStr)
}

// objectID -> string
func ObjectIDToString(id primitive.ObjectID) string {
	return id.Hex()
}
