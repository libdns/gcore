// Package gcore implements the libdns interfaces for GCore DNS
package gcore

import (
	"context"
	"fmt"
	"strings"
	"time"

	gcoreSDK "github.com/G-Core/gcore-dns-sdk-go"
	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with GCore DNS.
type Provider struct {
	APIKey string `json:"api_key,omitempty"`
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	cli := gcoreSDK.NewClient(gcoreSDK.PermanentAPIKeyAuth(p.APIKey))

	// Get records for zone and convert to libdns records
	gcoreZone, err := cli.Zone(ctx, zone)
	if err != nil {
		return nil, err
	}

	records := make([]libdns.Record, len(gcoreZone.Records))
	for i, gcoreRecord := range gcoreZone.Records {
		rrSets, err := cli.RRSet(ctx, zone, gcoreRecord.Name, gcoreRecord.Type)
		if err != nil {
			return nil, err
		}
		for _, rrSet := range rrSets.Records {
			records[i] = libdns.Record{
				Name:  gcoreRecord.Name,
				Type:  gcoreRecord.Type,
				TTL:   time.Duration(gcoreRecord.TTL) * time.Second,
				Value: rrSet.ContentToString(),
			}
		}
	}

	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	cli := gcoreSDK.NewClient(gcoreSDK.PermanentAPIKeyAuth(p.APIKey))

	// Validate records...
	// All records names must be fully qualified
	// CNAME records values must be fully qualified
	for _, record := range records {
		if record.Name == "@" {
			record.Name = zone
			continue
		}

		if !strings.HasSuffix(record.Name, "."+zone) {
			record.Name += "." + zone
		}

		if record.Type == "CNAME" && !strings.HasSuffix(record.Value, ".") {
			record.Value += "."
		}
	}

	recordsByType := make(map[string][]libdns.Record)
	for _, record := range records {
		recordsByType[record.Type] = append(recordsByType[record.Type], record)
	}

	var addedRecords []libdns.Record

	for recordType, records := range recordsByType {
		for _, record := range records {
			rrSet, err := cli.RRSet(ctx, zone, record.Name, recordType)
			if err != nil {
				if strings.Contains(err.Error(), "404: record is not found") {
					rrSet = gcoreSDK.RRSet{
						Type: recordType,
						TTL:  int(record.TTL.Seconds()),
						Records: []gcoreSDK.ResourceRecord{
							{
								Content: []any{record.Value},
								Enabled: true,
							},
						},
					}
					if err := cli.UpdateRRSet(ctx, zone, record.Name, recordType, rrSet); err != nil {
						return nil, err
					}
					addedRecords = append(addedRecords, record)
					continue
				}
				return nil, err
			}

			for _, rr := range rrSet.Records {
				if rr.ContentToString() == record.Value {
					continue
				}

				rrSet.Records = append(rrSet.Records, gcoreSDK.ResourceRecord{
					Content: []any{record.Value},
					Enabled: true,
				})
			}

			if err := cli.UpdateRRSet(ctx, zone, record.Name, recordType, rrSet); err != nil {
				return nil, err
			}
			addedRecords = append(addedRecords, record)
		}
	}

	return addedRecords, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	cli := gcoreSDK.NewClient(gcoreSDK.PermanentAPIKeyAuth(p.APIKey))

	var updatedRecords []libdns.Record

	for _, record := range records {
		rrSet, err := cli.RRSet(ctx, zone, record.Name, record.Type)
		if err != nil {
			return nil, err
		}

		for _, rr := range rrSet.Records {
			if rr.ContentToString() == record.Value {
				continue
			}

			rrSet.Records = append(rrSet.Records, gcoreSDK.ResourceRecord{
				Content: []any{record.Value},
				Enabled: true,
			})
		}

		if err := cli.UpdateRRSet(ctx, zone, record.Name, record.Type, rrSet); err != nil {
			return nil, err
		}

		updatedRecords = append(updatedRecords, record)
	}

	return updatedRecords, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	cli := gcoreSDK.NewClient(gcoreSDK.PermanentAPIKeyAuth(p.APIKey))

	var deletedRecords []libdns.Record

	for _, record := range records {
		if cli.DeleteRRSetRecord(ctx, zone, record.Name, record.Type, record.Value) != nil {
			return nil, fmt.Errorf("failed to delete record %v", record)
		}
		deletedRecords = append(deletedRecords, record)
	}

	return deletedRecords, nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
