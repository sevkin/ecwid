package ecwid

import (
	"context"
	"errors"
	"fmt"
	"html/template"
)

type (
	// NewCategory https://developers.ecwid.com/api-documentation/categories#add-new-category
	NewCategory struct {
		Name        string        `json:"name,omitempty"`
		ParentID    uint64        `json:"parentId"`
		OrderBy     int           `json:"orderBy"`
		Description template.HTML `json:"description,omitempty"`
		Enabled     bool          `json:"enabled"`
		ProductIDs  []uint64      `json:"productIds,omitempty"`
	}

	// Category https://developers.ecwid.com/api-documentation/categories#get-categories
	Category struct {
		NewCategory
		ID                  uint64        `json:"id"`
		HdThumbnailURL      string        `json:"hdThumbnailUrl"`
		ThumbnailURL        string        `json:"thumbnailUrl"`
		ImageURL            string        `json:"imageUrl"`
		OriginalImageURL    string        `json:"originalImageUrl"`
		URL                 string        `json:"url"`
		ProductCount        uint          `json:"productCount"`
		EnabledProductCount uint          `json:"enabledProductCount"`
		OriginalImage       *ImageDetails `json:"originalImage"`
	}

	// CategoriesGetResponse is basic details of found categories
	CategoriesGetResponse struct {
		Total  uint        `json:"total"`
		Count  uint        `json:"count"`
		Offset uint        `json:"offset"`
		Limit  uint        `json:"limit"`
		Items  []*Category `json:"items"`
	}
)

// CategoriesGet search or filter categories in a store catalog
// The response provides basic details of found categories
func (c *Client) CategoriesGet(filter map[string]string) (*CategoriesGetResponse, error) {
	// filter:
	// parent number, hidden_categories bool, offset number, limit number,
	// productIds array?, baseUrl string, cleanUrls bool

	response, err := c.R().
		SetQueryParams(filter).
		Get("/categories")

	var result CategoriesGetResponse
	return &result, responseUnmarshal(response, err, &result)
}

// Categories 'iterable' by filtered store categories
func (c *Client) Categories(ctx context.Context, filter map[string]string) <-chan *Category {
	catChan := make(chan *Category)

	go func() {
		defer close(catChan)

		c.CategoriesTrampoline(filter, func(index int, category *Category) error {
			// FIXME silent error. maybe catChan <- nil ?
			select {
			case <-ctx.Done():
				return errors.New("break")
			case catChan <- category:
			}
			return nil
		})
	}()

	return catChan
}

// CategoriesTrampoline call on each category
func (c *Client) CategoriesTrampoline(filter map[string]string, fn func(int, *Category) error) error {
	filterCopy := make(map[string]string)
	for k, v := range filter {
		filterCopy[k] = v
	}

	for {
		resp, err := c.CategoriesGet(filterCopy)
		if err != nil {
			return err
		}

		for index, category := range resp.Items {
			if err := fn(index, category); err != nil {
				return err
			}
		}

		if resp.Offset+resp.Count >= resp.Total {
			return nil
		}
		filterCopy["offset"] = fmt.Sprintf("%d", resp.Offset+resp.Count)
	}
}

// CategoryGet gets all details of a specific category in an Ecwid store by its ID
func (c *Client) CategoryGet(categoryID uint64) (*Category, error) {
	response, err := c.R().
		Get(fmt.Sprintf("/categories/%d", categoryID))

	var result Category
	return &result, responseUnmarshal(response, err, &result)
}

// CategoryAdd creates a new category in an Ecwid store
// returns new categoryId
func (c *Client) CategoryAdd(category *NewCategory) (uint64, error) {
	response, err := c.R().
		SetHeader("Content-Type", "application/json").
		SetBody(category).
		Post("/categories")

	return responseAdd(response, err)
}

// CategoryUpdate update an existing category in an Ecwid store referring to its ID
func (c *Client) CategoryUpdate(categoryID uint64, category *NewCategory) error {
	response, err := c.R().
		SetHeader("Content-Type", "application/json").
		SetBody(category).
		Put(fmt.Sprintf("/categories/%d", categoryID))

	return responseUpdate(response, err)
}

// CategoryDelete delete a category from an Ecwid store referring to its ID
func (c *Client) CategoryDelete(categoryID uint64) error {
	response, err := c.R().
		Delete(fmt.Sprintf("/categories/%d", categoryID))

	_, err = responseDelete(response, err)
	return err
}
