package services

import (
	"context"
	"go-grpc/cmd/helpers"
	pagingPb "go-grpc/pb/pagination"
	productPb "go-grpc/pb/product"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type ProductService struct {
	productPb.UnimplementedProductServiceServer
	DB *gorm.DB
}

func (p *ProductService) GetProducts(ctx context.Context, pageParam *productPb.Page) (*productPb.Products, error) {
	var page int64 = 1

	if pageParam.GetPage() != 0 {
		page = int64(pageParam.GetPage())
	}

	var pagination pagingPb.Pagination
	var products []*productPb.Product

	sql := p.DB.Table("products AS p").Joins("JOIN categories AS c ON p.category_id = c.id").Select("p.id", "p.name", "p.price", "p.stock", "c.name AS category_name")

	offset, limit := helpers.Pagination(sql, page, &pagination)
	rows, err := sql.Offset(int(offset)).Limit(int(limit)).Rows()

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	defer rows.Close()

	for rows.Next() {
		var product productPb.Product
		var category productPb.Category

		if err := rows.Scan(&product.Id, &product.Name, &product.Price, &product.Stock, &category.Name); err != nil {
			log.Fatalf("Error scanning rows: %v", err.Error())
		}

		product.Category = &category
		products = append(products, &product)
	}

	response := &productPb.Products{
		Pagination: &pagination,
		Data:       products,
	}

	return response, nil
}

func (p *ProductService) GetProduct(ctx context.Context, idParam *productPb.Id) (*productPb.Product, error) {
	sql := p.DB.Table("products AS p").Joins("JOIN categories AS c ON p.category_id = c.id").Select("p.id", "p.name", "p.price", "p.stock", "c.name AS category_name").Where("p.id = ?", idParam.GetId()).Row()

	var product productPb.Product
	var category productPb.Category

	if err := sql.Scan(&product.Id, &product.Name, &product.Price, &product.Stock, &category.Name); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	product.Category = &category

	return &product, nil
}

func (p *ProductService) CreateProduct(ctx context.Context, productData *productPb.Product) (*productPb.Id, error) {
	var Response productPb.Id

	err := p.DB.Transaction(func(tx *gorm.DB) error {
		category := productPb.Category{
			Id:   0,
			Name: productData.GetCategory().GetName(),
		}

		if err := tx.Table("categories").Where("LCASE(name) = ?", category.GetName()).FirstOrCreate(&category).Error; err != nil {
			return err
		}

		product := struct {
			Id          uint64
			Name        string
			Price       float64
			Stock       uint32
			Category_id uint32
		}{
			Id:          productData.GetId(),
			Name:        productData.GetName(),
			Price:       productData.GetPrice(),
			Stock:       productData.GetStock(),
			Category_id: category.GetId(),
		}

		if err := tx.Table("products").Create(&product).Error; err != nil {
			return err
		}

		Response.Id = product.Id

		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &Response, nil
}

func (p *ProductService) UpdateProduct(ctx context.Context, productData *productPb.Product) (*productPb.Status, error) {
	var Response productPb.Status

	err := p.DB.Transaction(func(tx *gorm.DB) error {
		category := productPb.Category{
			Id:   0,
			Name: productData.GetCategory().GetName(),
		}

		if err := tx.Table("categories").Where("LCASE(name) = ?", category.GetName()).FirstOrCreate(&category).Error; err != nil {
			return err
		}

		product := struct {
			Id          uint64
			Name        string
			Price       float64
			Stock       uint32
			Category_id uint32
		}{
			Id:          productData.GetId(),
			Name:        productData.GetName(),
			Price:       productData.GetPrice(),
			Stock:       productData.GetStock(),
			Category_id: category.GetId(),
		}

		if err := tx.Table("products").Where("id = ?", product.Id).Updates(&product).Error; err != nil {
			return err
		}

		Response.Status = 1

		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &Response, nil
}

func (p *ProductService) DeleteProduct(ctx context.Context, idParam *productPb.Id) (*productPb.Status, error) {
	var Response productPb.Status

	if err := p.DB.Table("products").Where("id = ?", idParam.GetId()).Delete(nil).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	Response.Status = 1

	return &Response, nil
}
