package modbusgw

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"time"
)

func (m *ModbusGateway) Gather(acc telegraf.Accumulator) error {
	if !m.isConnected {
		err := connect(m)
		if err != nil {
			m.isConnected = false
			return err
		}
	}

	grouper := metric.NewSeriesGrouper()

	for _, req := range m.Requests {
		now := time.Now()
		m.tcpHandler.SlaveId = req.Unit
		var resp []byte
		var err error

		if req.Type == "holding" {
			resp, err = m.client.ReadHoldingRegisters(req.Address, req.Count)
		} else if req.Type == "input" {
			resp, err = m.client.ReadInputRegisters(req.Address, req.Count)
		} else {
			return fmt.Errorf("Don't know how to poll register type \"%s\"", req.Type)
		}

		if err == nil {

			reader := bytes.NewReader(resp)

			for _, f := range req.Fields {
				switch f.Type {
				case "UINT16":
					var value uint16
					binary.Read(reader, binary.BigEndian, &value)
					writeInt(grouper, &req, &f, int64(value), now)
					break
				case "INT16":
					var value int16
					binary.Read(reader, binary.BigEndian, &value)
					writeInt(grouper, &req, &f, int64(value), now)
					break
				case "UINT32":
					var value uint32
					binary.Read(reader, binary.BigEndian, &value)
					writeInt(grouper, &req, &f, int64(value), now)
					break
				case "INT32":
					var value int32
					binary.Read(reader, binary.BigEndian, &value)
					writeInt(grouper, &req, &f, int64(value), now)

					break

				}

			}

		} else {

			m.Log.Info("Modbus Error: ", err)
		}

	}

	for _, metric := range grouper.Metrics() {
		acc.AddMetric(metric)
	}

	return nil
}