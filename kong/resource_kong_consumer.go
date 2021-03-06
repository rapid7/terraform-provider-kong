package kong

import (
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
	"github.com/hashicorp/terraform/helper/schema"
)

type Consumer struct {
	ID       string `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
	CustomID string `json:"custom_id,omitempty"`
}

func resourceKongConsumer() *schema.Resource {
	return &schema.Resource{
		Create: resourceKongConsumerCreate,
		Read:   resourceKongConsumerRead,
		Update: resourceKongConsumerUpdate,
		Delete: resourceKongConsumerDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     nil,
				Description: "The username of the consumer. You must send either this field or custom_id with the request.",
			},

			"custom_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     nil,
				Description: "Field for storing an existing ID for the consumer, useful for mapping Kong with users in your existing database. You must send either this field or username with the request.",
			},
		},
	}
}

func resourceKongConsumerCreate(d *schema.ResourceData, meta interface{}) error {
	sling := meta.(*sling.Sling)

	consumer := getConsumerFromResourceData(d)

	createdConsumer := new(Consumer)

	errorResponse := make(map[string]interface{})
	response, error := sling.New().BodyJSON(consumer).Post("consumers/").Receive(createdConsumer, &errorResponse)
	if error != nil {
		return fmt.Errorf("Error while creating consumer.")
	}

	if response.StatusCode == http.StatusConflict {
		return fmt.Errorf("409 Conflict - use terraform import to manage this consumer.")
	} else if response.StatusCode != http.StatusCreated {
		return ErrorFromResponse(response, errorResponse)
	}

	setConsumerToResourceData(d, createdConsumer)

	return nil
}

func resourceKongConsumerRead(d *schema.ResourceData, meta interface{}) error {
	sling := meta.(*sling.Sling)

	id := d.Id()
	consumer := new(Consumer)

	errorResponse := make(map[string]interface{})
	response, error := sling.New().Path("consumers/").Get(id).Receive(consumer, &errorResponse)
	if error != nil {
		return fmt.Errorf("Error while updating consumer.")
	}

	if response.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	} else if response.StatusCode != http.StatusOK {
		return ErrorFromResponse(response, errorResponse)
	}

	setConsumerToResourceData(d, consumer)

	return nil
}

func resourceKongConsumerUpdate(d *schema.ResourceData, meta interface{}) error {
	sling := meta.(*sling.Sling)

	consumer := getConsumerFromResourceData(d)

	updatedConsumer := new(Consumer)

	errorResponse := make(map[string]interface{})
	response, error := sling.New().BodyJSON(consumer).Patch("consumers/").Path(consumer.ID).Receive(updatedConsumer, &errorResponse)
	if error != nil {
		return fmt.Errorf("Error while updating consumer.")
	}

	if response.StatusCode != http.StatusOK {
		return ErrorFromResponse(response, errorResponse)
	}

	setConsumerToResourceData(d, updatedConsumer)

	return nil
}

func resourceKongConsumerDelete(d *schema.ResourceData, meta interface{}) error {
	sling := meta.(*sling.Sling)

	id := d.Id()

	errorResponse := make(map[string]interface{})
	response, error := sling.New().Delete("consumers/").Path(id).Receive(nil, &errorResponse)
	if error != nil {
		return fmt.Errorf("Error while deleting consumer.")
	}

	if response.StatusCode != http.StatusNoContent {
		return ErrorFromResponse(response, errorResponse)
	}

	return nil
}

func getConsumerFromResourceData(d *schema.ResourceData) *Consumer {
	consumer := &Consumer{
		ID:       d.Id(),
		Username: d.Get("username").(string),
		CustomID: d.Get("custom_id").(string),
	}

	return consumer
}

func setConsumerToResourceData(d *schema.ResourceData, consumer *Consumer) {
	d.SetId(consumer.ID)
	d.Set("username", consumer.Username)
	d.Set("custom_id", consumer.CustomID)
}
