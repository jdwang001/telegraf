package modbusgw

import (
	"fmt"
	"strings"
)

func (m *ModbusGateway) Init() error {

	for i := range m.Requests {
		request := &m.Requests[i]
		/*
		 * Check register type is valid
		 */
		if request.Type == "" {
			request.Type = "holding"
		}

		request.Type = strings.ToLower(request.Type)
		if request.Type != "holding" && request.Type != "input" {
			return fmt.Errorf("Request type must be \"holding\" or \"input\"")
		}

		/*
		 * Check field mappings
		 */
		for j := range m.Requests[i].Fields {
			field := &m.Requests[i].Fields[j]

			if field.Scale == 0.0 {
				field.Scale = 1.0
			}

		}
	}

	return nil
}