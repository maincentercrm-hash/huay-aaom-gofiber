package controllers

import (
	"context"
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GenericController struct {
	Collection   *mongo.Collection
	SearchFields []string
	SortFields   []string
}

func NewGenericController(collection *mongo.Collection, searchFields []string, sortFields []string) *GenericController {
	return &GenericController{
		Collection:   collection,
		SearchFields: searchFields,
		SortFields:   sortFields,
	}
}

func (gc *GenericController) Create(c *fiber.Ctx) error {
	var data map[string]interface{}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	result, err := gc.Collection.InsertOne(context.Background(), data)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create item"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": result.InsertedID})
}

func (gc *GenericController) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	skip := (page - 1) * limit

	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit))
	cursor, err := gc.Collection.Find(context.Background(), bson.M{}, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch items"})
	}
	defer cursor.Close(context.Background())

	var items []bson.M
	if err = cursor.All(context.Background(), &items); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode items"})
	}

	totalItems, _ := gc.Collection.CountDocuments(context.Background(), bson.M{})
	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

	return c.JSON(fiber.Map{
		"items":       items,
		"currentPage": page,
		"totalPages":  totalPages,
		"totalItems":  totalItems,
	})
}

func (gc *GenericController) GetById(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var item bson.M
	err = gc.Collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&item)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Item not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch item"})
	}

	return c.JSON(item)
}

func (gc *GenericController) Update(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var data map[string]interface{}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	update := bson.M{"$set": data}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedItem bson.M
	err = gc.Collection.FindOneAndUpdate(context.Background(), bson.M{"_id": id}, update, opts).Decode(&updatedItem)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Item not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update item"})
	}

	return c.JSON(updatedItem)
}

func (gc *GenericController) Delete(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	result, err := gc.Collection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete item"})
	}

	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Item not found"})
	}

	return c.JSON(fiber.Map{"message": "Item deleted successfully"})
}

func (gc *GenericController) Search(c *fiber.Ctx) error {
	query := c.Query("query")
	field := c.Query("field", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	skip := (page - 1) * limit

	searchQuery := bson.M{}
	if query != "" && field != "" {
		searchQuery[field] = primitive.Regex{Pattern: query, Options: "i"}
	} else if query != "" {
		orConditions := []bson.M{}
		for _, field := range gc.SearchFields {
			orConditions = append(orConditions, bson.M{field: primitive.Regex{Pattern: query, Options: "i"}})
		}
		searchQuery["$or"] = orConditions
	}

	// Create dynamic sort based on SortFields
	sort := bson.D{}
	for _, field := range gc.SortFields {
		// Assume descending order for all fields. You can modify this logic if needed.
		sort = append(sort, bson.E{Key: field, Value: -1})
	}

	opts := options.Find().
		SetSort(sort).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := gc.Collection.Find(context.Background(), searchQuery, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to search items"})
	}
	defer cursor.Close(context.Background())

	var items []bson.M
	if err = cursor.All(context.Background(), &items); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode items"})
	}

	totalItems, _ := gc.Collection.CountDocuments(context.Background(), searchQuery)
	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

	return c.JSON(fiber.Map{
		"items":       items,
		"currentPage": page,
		"totalPages":  totalPages,
		"totalItems":  totalItems,
	})
}
