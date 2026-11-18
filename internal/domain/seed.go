package domain

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"
)

func (s *Store) seedDemoData() {
	baseTime := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)

	// 1. Seed Customers (CUST-1001 .. CUST-1012)
	customers := []Customer{
		{
			ID: "CUST-1001", Name: "Acme Industrial Systems", LegalName: "Acme Industrial Systems Inc.",
			TaxID: "US-94-3829104", Industry: "Manufacturing", SalesOrg: "1000", Currency: "USD",
			PaymentTerms: "NET30", CreditLimit: 250000, CreditUsed: 38880, Status: "ACTIVE",
			Country: "US", City: "Chicago", Street: "450 W Randolph St", PostalCode: "60606",
			Email: "ap@acme-industrial.example.com", Phone: "+1-312-555-0140", CreatedAt: baseTime,
		},
		{
			ID: "CUST-1002", Name: "RheinMetallurgie GmbH", LegalName: "RheinMetallurgie GmbH & Co. KG",
			TaxID: "DE-284910382", Industry: "Automotive", SalesOrg: "2000", Currency: "EUR",
			PaymentTerms: "NET45", CreditLimit: 400000, CreditUsed: 51840, Status: "ACTIVE",
			Country: "DE", City: "Stuttgart", Street: "Industriestrasse 42", PostalCode: "70565",
			Email: "einkauf@rheinmetallurgie.example.de", Phone: "+49-711-555-0192", CreatedAt: baseTime,
		},
		{
			ID: "CUST-1003", Name: "Kuroda Precision Robotics", LegalName: "Kuroda Precision Robotics KK",
			TaxID: "JP-883920194", Industry: "Robotics", SalesOrg: "3000", Currency: "USD",
			PaymentTerms: "NET30", CreditLimit: 300000, CreditUsed: 21600, Status: "ACTIVE",
			Country: "JP", City: "Osaka", Street: "2-4-12 Namba", PostalCode: "542-0076",
			Email: "procurement@kuroda-robotics.example.jp", Phone: "+81-6-5550-0188", CreatedAt: baseTime,
		},
		{
			ID: "CUST-1004", Name: "Thames Valley Logistics", LegalName: "Thames Valley Logistics Ltd",
			TaxID: "GB-774829102", Industry: "Logistics", SalesOrg: "2000", Currency: "GBP",
			PaymentTerms: "NET30", CreditLimit: 150000, CreditUsed: 16200, Status: "ACTIVE",
			Country: "GB", City: "Reading", Street: "180 Thameside Way", PostalCode: "RG1 8EX",
			Email: "finance@thameslogistics.example.co.uk", Phone: "+44-118-555-0123", CreatedAt: baseTime,
		},
		{
			ID: "CUST-1005", Name: "Solstice Energy Grid", LegalName: "Solstice Energy Grid Corp.",
			TaxID: "US-77-1920384", Industry: "Utilities", SalesOrg: "1000", Currency: "USD",
			PaymentTerms: "NET60", CreditLimit: 500000, CreditUsed: 91800, Status: "ACTIVE",
			Country: "US", City: "Austin", Street: "900 Congress Ave", PostalCode: "78701",
			Email: "billing@solstice-energy.example.com", Phone: "+1-512-555-0177", CreatedAt: baseTime,
		},
		{
			ID: "CUST-1006", Name: "Nordic ColdChain AB", LegalName: "Nordic ColdChain AB",
			TaxID: "SE-5568291042", Industry: "Cold Logistics", SalesOrg: "2000", Currency: "EUR",
			PaymentTerms: "NET30", CreditLimit: 180000, CreditUsed: 0, Status: "ACTIVE",
			Country: "SE", City: "Gothenburg", Street: "Hamngatan 15", PostalCode: "411 14",
			Email: "orders@nordiccoldchain.example.se", Phone: "+46-31-555-0110", CreatedAt: baseTime,
		},
		{
			ID: "CUST-1007", Name: "Atelier Lumiere SAS", LegalName: "Atelier Lumiere SAS",
			TaxID: "FR-44829103849", Industry: "Electronics", SalesOrg: "2000", Currency: "EUR",
			PaymentTerms: "2%10NET30", CreditLimit: 120000, CreditUsed: 11340, Status: "ACTIVE",
			Country: "FR", City: "Lyon", Street: "24 Rue de la Republique", PostalCode: "69002",
			Email: "comptabilite@atelier-lumiere.example.fr", Phone: "+33-4-72-55-0145", CreatedAt: baseTime,
		},
		{
			ID: "CUST-1008", Name: "Pacific Rim Semiconductors", LegalName: "Pacific Rim Semiconductors Pte Ltd",
			TaxID: "SG-201948271K", Industry: "Semiconductors", SalesOrg: "3000", Currency: "USD",
			PaymentTerms: "NET45", CreditLimit: 600000, CreditUsed: 73440, Status: "ACTIVE",
			Country: "SG", City: "Singapore", Street: "10 Marina Blvd", PostalCode: "018983",
			Email: "ap@pacificrim-semi.example.sg", Phone: "+65-6555-0199", CreatedAt: baseTime,
		},
		{
			ID: "CUST-1009", Name: "Maple Leaf Aerospace", LegalName: "Maple Leaf Aerospace Inc.",
			TaxID: "CA-839201948", Industry: "Aerospace", SalesOrg: "1000", Currency: "USD",
			PaymentTerms: "NET30", CreditLimit: 350000, CreditUsed: 27000, Status: "ACTIVE",
			Country: "CA", City: "Montreal", Street: "1200 Rene-Levesque", PostalCode: "H3B 4W8",
			Email: "supply@mapleleaf-aero.example.ca", Phone: "+1-514-555-0162", CreatedAt: baseTime,
		},
		{
			ID: "CUST-1010", Name: "Vanguard Retail Holdings", LegalName: "Vanguard Retail Holdings LLC",
			TaxID: "US-61-8291034", Industry: "Retail", SalesOrg: "1000", Currency: "USD",
			PaymentTerms: "NET30", CreditLimit: 75000, CreditUsed: 74200, Status: "BLOCKED",
			Country: "US", City: "Atlanta", Street: "200 Peachtree St", PostalCode: "30303",
			Email: "ap@vanguard-retail.example.com", Phone: "+1-404-555-0181", CreatedAt: baseTime,
		},
		{
			ID: "CUST-1011", Name: "Delta Port Automation", LegalName: "Delta Port Automation BV",
			TaxID: "NL-829103847B01", Industry: "Maritime", SalesOrg: "2000", Currency: "EUR",
			PaymentTerms: "NET45", CreditLimit: 220000, CreditUsed: 19440, Status: "ACTIVE",
			Country: "NL", City: "Rotterdam", Street: "Wilhelminakade 173", PostalCode: "3072 AP",
			Email: "finance@deltaport.example.nl", Phone: "+31-10-555-0134", CreatedAt: baseTime,
		},
		{
			ID: "CUST-1012", Name: "Sierra Mining & Minerals", LegalName: "Sierra Mining & Minerals SA",
			TaxID: "US-88-4719203", Industry: "Mining", SalesOrg: "1000", Currency: "USD",
			PaymentTerms: "NET30", CreditLimit: 200000, CreditUsed: 0, Status: "INACTIVE",
			Country: "US", City: "Denver", Street: "1600 Broadway", PostalCode: "80202",
			Email: "accounts@sierramining.example.com", Phone: "+1-303-555-0155", CreatedAt: baseTime,
		},
	}
	for i := range customers {
		c := customers[i]
		s.customers[c.ID] = &c
	}
	s.nextCustomerSeq = 1013

	// 2. Seed Materials (MAT-1001 .. MAT-1016)
	materials := []Material{
		{ID: "MAT-1001", SKU: "SRV-EDGE-X4", Name: "Industrial Edge Gateway X4", Description: "Ruggedized DIN-rail IIoT edge compute node with dual 10GbE", Type: "FERT", Category: "Edge Compute", BaseUOM: "EA", UnitPrice: 1850.00, Currency: "USD", StockQty: 140, ReorderPoint: 25, Plant: "1000", StorageLocation: "0001", Status: "ACTIVE"},
		{ID: "MAT-1002", SKU: "PLC-S7-900", Name: "Modular PLC Controller Pro", Description: "High-speed programmable logic controller for factory automation", Type: "FERT", Category: "Automation", BaseUOM: "EA", UnitPrice: 3200.00, Currency: "USD", StockQty: 65, ReorderPoint: 15, Plant: "1000", StorageLocation: "0001", Status: "ACTIVE"},
		{ID: "MAT-1003", SKU: "SNS-VIB-200", Name: "Tri-Axial Vibration Sensor", Description: "Wireless predictive maintenance sensor for rotating machinery", Type: "FERT", Category: "Sensors", BaseUOM: "EA", UnitPrice: 240.00, Currency: "USD", StockQty: 420, ReorderPoint: 80, Plant: "1000", StorageLocation: "0002", Status: "ACTIVE"},
		{ID: "MAT-1004", SKU: "HMI-TCH-15", Name: "15-inch Industrial Touch Panel", Description: "IP65-rated capacitive HMI display with PROFINET interface", Type: "FERT", Category: "HMI", BaseUOM: "EA", UnitPrice: 1450.00, Currency: "USD", StockQty: 18, ReorderPoint: 20, Plant: "2000", StorageLocation: "0001", Status: "ACTIVE"}, // Low stock!
		{ID: "MAT-1005", SKU: "PSU-24V-40A", Name: "24V DC 40A Redundant Power Supply", Description: "DIN-rail industrial power supply with active load sharing", Type: "HAWA", Category: "Power", BaseUOM: "EA", UnitPrice: 380.00, Currency: "USD", StockQty: 210, ReorderPoint: 40, Plant: "1000", StorageLocation: "0002", Status: "ACTIVE"},
		{ID: "MAT-1006", SKU: "SW-SCADA-ENT", Name: "SCADA Enterprise Site License", Description: "Annual per-site telemetry and supervisory control software license", Type: "FERT", Category: "Software", BaseUOM: "EA", UnitPrice: 12000.00, Currency: "USD", StockQty: 999, ReorderPoint: 10, Plant: "1000", StorageLocation: "0001", Status: "ACTIVE"},
		{ID: "MAT-1007", SKU: "RACK-42U-IP55", Name: "Sealed 42U Factory Server Cabinet", Description: "Dust and splash-proof industrial enclosure with active cooling", Type: "FERT", Category: "Enclosures", BaseUOM: "EA", UnitPrice: 4500.00, Currency: "USD", StockQty: 12, ReorderPoint: 15, Plant: "2000", StorageLocation: "0001", Status: "ACTIVE"}, // Low stock!
		{ID: "MAT-1008", SKU: "CAB-FO-100M", Name: "Armored Fiber Optic Trunk 100m", Description: "OM4 multimode armored industrial fiber optic cable reel", Type: "HAWA", Category: "Cabling", BaseUOM: "BOX", UnitPrice: 620.00, Currency: "USD", StockQty: 95, ReorderPoint: 20, Plant: "1000", StorageLocation: "0002", Status: "ACTIVE"},
		{ID: "MAT-1009", SKU: "UPS-3KVA-LI", Name: "3kVA Lithium Rackmount UPS", Description: "Online double-conversion uninterruptible power supply", Type: "HAWA", Category: "Power", BaseUOM: "EA", UnitPrice: 2100.00, Currency: "USD", StockQty: 34, ReorderPoint: 10, Plant: "2000", StorageLocation: "0001", Status: "ACTIVE"},
		{ID: "MAT-1010", SKU: "CAM-VIS-4K", Name: "4K Machine Vision Inspection Camera", Description: "Global-shutter PoE+ industrial camera for defect detection", Type: "FERT", Category: "Vision", BaseUOM: "EA", UnitPrice: 1950.00, Currency: "USD", StockQty: 48, ReorderPoint: 12, Plant: "2000", StorageLocation: "0001", Status: "ACTIVE"},
		{ID: "MAT-1011", SKU: "DRV-SRV-5KW", Name: "5kW AC Servo Drive Unit", Description: "Closed-loop motion control drive with EtherCAT support", Type: "FERT", Category: "Motion", BaseUOM: "EA", UnitPrice: 1680.00, Currency: "USD", StockQty: 55, ReorderPoint: 15, Plant: "1000", StorageLocation: "0001", Status: "ACTIVE"},
		{ID: "MAT-1012", SKU: "SNS-THM-IR", Name: "Infrared Thermal Array Sensor", Description: "Non-contact temperature monitoring array for kilns & switchgear", Type: "FERT", Category: "Sensors", BaseUOM: "EA", UnitPrice: 890.00, Currency: "USD", StockQty: 7, ReorderPoint: 15, Plant: "1000", StorageLocation: "0002", Status: "ACTIVE"}, // Low stock!
		{ID: "MAT-1013", SKU: "RAW-ALU-6061", Name: "6061-T6 Extruded Aluminum Profile", Description: "Anodized structural framing bar, 3m length", Type: "ROH", Category: "Raw Material", BaseUOM: "EA", UnitPrice: 85.00, Currency: "USD", StockQty: 600, ReorderPoint: 100, Plant: "1000", StorageLocation: "0003", Status: "ACTIVE"},
		{ID: "MAT-1014", SKU: "RAW-PCB-8L", Name: "8-Layer FR4 Bare PCB Panel", Description: "High-Tg lead-free circuit board substrate for controller assembly", Type: "ROH", Category: "Raw Material", BaseUOM: "SET", UnitPrice: 140.00, Currency: "USD", StockQty: 320, ReorderPoint: 75, Plant: "2000", StorageLocation: "0003", Status: "ACTIVE"},
		{ID: "MAT-1015", SKU: "KIT-DIN-MNT", Name: "Heavy-Duty DIN Rail Mounting Kit", Description: "Stainless steel bracket and grounding terminal assortment", Type: "HAWA", Category: "Accessories", BaseUOM: "BOX", UnitPrice: 45.00, Currency: "USD", StockQty: 850, ReorderPoint: 150, Plant: "1000", StorageLocation: "0002", Status: "ACTIVE"},
		{ID: "MAT-1016", SKU: "MOD-5G-IND", Name: "Private 5G Industrial Modem", Description: "Dual-SIM 5G NR / LTE Cat-19 industrial cellular router", Type: "FERT", Category: "Networking", BaseUOM: "EA", UnitPrice: 1120.00, Currency: "USD", StockQty: 82, ReorderPoint: 20, Plant: "2000", StorageLocation: "0001", Status: "ACTIVE"},
	}
	for i := range materials {
		m := materials[i]
		s.materials[m.ID] = &m
	}
	s.nextMaterialSeq = 1017

	// 3. Seed Vendors (VEND-1001 .. VEND-1008)
	vendors := []Vendor{
		{ID: "VEND-1001", Name: "SiliconFab Precision Corp", Category: "Semiconductors & PCBs", Country: "TW", City: "Hsinchu", Currency: "USD", PaymentTerms: "NET45", Rating: "A", Email: "orders@siliconfab.example.tw", TaxID: "TW-28491032", Status: "ACTIVE"},
		{ID: "VEND-1002", Name: "Bavaria Stahl & Alu GmbH", Category: "Metals & Enclosures", Country: "DE", City: "Munich", Currency: "EUR", PaymentTerms: "NET30", Rating: "A", Email: "vertrieb@bavariastahl.example.de", TaxID: "DE-192837465", Status: "ACTIVE"},
		{ID: "VEND-1003", Name: "OptiLens Photonics Ltd", Category: "Optics & Vision", Country: "JP", City: "Kyoto", Currency: "USD", PaymentTerms: "NET30", Rating: "A", Email: "sales@optilens.example.jp", TaxID: "JP-449102834", Status: "ACTIVE"},
		{ID: "VEND-1004", Name: "Midwest Power Components", Category: "Power Electronics", Country: "US", City: "Milwaukee", Currency: "USD", PaymentTerms: "NET30", Rating: "B", Email: "ar@midwestpower.example.com", TaxID: "US-39-4829104", Status: "ACTIVE"},
		{ID: "VEND-1005", Name: "Helvetia Motion Systems", Category: "Servo & Drives", Country: "CH", City: "Zurich", Currency: "EUR", PaymentTerms: "NET60", Rating: "A", Email: "supply@helvetiamotion.example.ch", TaxID: "CHE-492.103.847", Status: "ACTIVE"},
		{ID: "VEND-1006", Name: "Shenzhen FiberLink Co", Category: "Cabling & Connectors", Country: "CN", City: "Shenzhen", Currency: "USD", PaymentTerms: "NET30", Rating: "B", Email: "export@fiberlink.example.cn", TaxID: "CN-91440300", Status: "ACTIVE"},
		{ID: "VEND-1007", Name: "GreatLakes Packaging & Crating", Category: "Packaging", Country: "US", City: "Detroit", Currency: "USD", PaymentTerms: "NET15", Rating: "B", Email: "orders@glpackaging.example.com", TaxID: "US-38-1928374", Status: "ACTIVE"},
		{ID: "VEND-1008", Name: "Iberia Die-Cast Holdings", Category: "Castings", Country: "ES", City: "Bilbao", Currency: "EUR", PaymentTerms: "NET45", Rating: "C", Email: "sales@iberiadiecast.example.es", TaxID: "ES-B48291038", Status: "ON_HOLD"},
	}
	for i := range vendors {
		v := vendors[i]
		s.vendors[v.ID] = &v
	}
	s.nextVendorSeq = 1009

	// 4. Seed Sales Orders (SO-50001 .. SO-50010)
	salesOrders := []SalesOrder{
		{
			ID: "SO-50001", CustomerID: "CUST-1001", CustomerName: "Acme Industrial Systems", SalesOrg: "1000",
			OrderDate: "2026-10-05", RequestedDeliveryDate: "2026-10-15", Status: "BILLED", Currency: "USD",
			NetAmount: 36000, TaxAmount: 2880, TotalAmount: 38880, ReferencePO: "ACME-PO-8812",
			Items: []SalesOrderItem{
				{ItemNumber: 10, MaterialID: "MAT-1001", MaterialName: "Industrial Edge Gateway X4", Quantity: 10, UOM: "EA", UnitPrice: 1850, NetAmount: 18500},
				{ItemNumber: 20, MaterialID: "MAT-1002", MaterialName: "Modular PLC Controller Pro", Quantity: 5, UOM: "EA", UnitPrice: 3200, NetAmount: 16000},
				{ItemNumber: 30, MaterialID: "MAT-1003", MaterialName: "Tri-Axial Vibration Sensor", Quantity: 6, UOM: "EA", UnitPrice: 250, NetAmount: 1500},
			},
			CreatedAt: baseTime.AddDate(0, 0, 4),
		},
		{
			ID: "SO-50002", CustomerID: "CUST-1002", CustomerName: "RheinMetallurgie GmbH", SalesOrg: "2000",
			OrderDate: "2026-10-10", RequestedDeliveryDate: "2026-10-24", Status: "BILLED", Currency: "EUR",
			NetAmount: 48000, TaxAmount: 3840, TotalAmount: 51840, ReferencePO: "RM-2026-441",
			Items: []SalesOrderItem{
				{ItemNumber: 10, MaterialID: "MAT-1006", MaterialName: "SCADA Enterprise Site License", Quantity: 2, UOM: "EA", UnitPrice: 12000, NetAmount: 24000},
				{ItemNumber: 20, MaterialID: "MAT-1002", MaterialName: "Modular PLC Controller Pro", Quantity: 6, UOM: "EA", UnitPrice: 3200, NetAmount: 19200},
				{ItemNumber: 30, MaterialID: "MAT-1003", MaterialName: "Tri-Axial Vibration Sensor", Quantity: 20, UOM: "EA", UnitPrice: 240, NetAmount: 4800},
			},
			CreatedAt: baseTime.AddDate(0, 0, 9),
		},
		{
			ID: "SO-50003", CustomerID: "CUST-1003", CustomerName: "Kuroda Precision Robotics", SalesOrg: "3000",
			OrderDate: "2026-10-14", RequestedDeliveryDate: "2026-10-28", Status: "DELIVERED", Currency: "USD",
			NetAmount: 20000, TaxAmount: 1600, TotalAmount: 21600, ReferencePO: "KPR-9920",
			Items: []SalesOrderItem{
				{ItemNumber: 10, MaterialID: "MAT-1010", MaterialName: "4K Machine Vision Inspection Camera", Quantity: 6, UOM: "EA", UnitPrice: 1950, NetAmount: 11700},
				{ItemNumber: 20, MaterialID: "MAT-1011", MaterialName: "5kW AC Servo Drive Unit", Quantity: 5, UOM: "EA", UnitPrice: 1660, NetAmount: 8300},
			},
			CreatedAt: baseTime.AddDate(0, 0, 13),
		},
		{
			ID: "SO-50004", CustomerID: "CUST-1005", CustomerName: "Solstice Energy Grid", SalesOrg: "1000",
			OrderDate: "2026-10-18", RequestedDeliveryDate: "2026-11-01", Status: "BILLED", Currency: "USD",
			NetAmount: 85000, TaxAmount: 6800, TotalAmount: 91800, ReferencePO: "SOL-GRID-109",
			Items: []SalesOrderItem{
				{ItemNumber: 10, MaterialID: "MAT-1006", MaterialName: "SCADA Enterprise Site License", Quantity: 4, UOM: "EA", UnitPrice: 12000, NetAmount: 48000},
				{ItemNumber: 20, MaterialID: "MAT-1001", MaterialName: "Industrial Edge Gateway X4", Quantity: 20, UOM: "EA", UnitPrice: 1850, NetAmount: 37000},
			},
			CreatedAt: baseTime.AddDate(0, 0, 17),
		},
		{
			ID: "SO-50005", CustomerID: "CUST-1004", CustomerName: "Thames Valley Logistics", SalesOrg: "2000",
			OrderDate: "2026-10-22", RequestedDeliveryDate: "2026-11-05", Status: "CONFIRMED", Currency: "GBP",
			NetAmount: 15000, TaxAmount: 1200, TotalAmount: 16200, ReferencePO: "TVL-771",
			Items: []SalesOrderItem{
				{ItemNumber: 10, MaterialID: "MAT-1016", MaterialName: "Private 5G Industrial Modem", Quantity: 10, UOM: "EA", UnitPrice: 1120, NetAmount: 11200},
				{ItemNumber: 20, MaterialID: "MAT-1005", MaterialName: "24V DC 40A Redundant Power Supply", Quantity: 10, UOM: "EA", UnitPrice: 380, NetAmount: 3800},
			},
			CreatedAt: baseTime.AddDate(0, 0, 21),
		},
		{
			ID: "SO-50006", CustomerID: "CUST-1008", CustomerName: "Pacific Rim Semiconductors", SalesOrg: "3000",
			OrderDate: "2026-10-28", RequestedDeliveryDate: "2026-11-12", Status: "BILLED", Currency: "USD",
			NetAmount: 68000, TaxAmount: 5440, TotalAmount: 73440, ReferencePO: "PRS-FAB-22",
			Items: []SalesOrderItem{
				{ItemNumber: 10, MaterialID: "MAT-1010", MaterialName: "4K Machine Vision Inspection Camera", Quantity: 20, UOM: "EA", UnitPrice: 1950, NetAmount: 39000},
				{ItemNumber: 20, MaterialID: "MAT-1004", MaterialName: "15-inch Industrial Touch Panel", Quantity: 20, UOM: "EA", UnitPrice: 1450, NetAmount: 29000},
			},
			CreatedAt: baseTime.AddDate(0, 0, 27),
		},
		{
			ID: "SO-50007", CustomerID: "CUST-1007", CustomerName: "Atelier Lumiere SAS", SalesOrg: "2000",
			OrderDate: "2026-11-02", RequestedDeliveryDate: "2026-11-18", Status: "OPEN", Currency: "EUR",
			NetAmount: 10500, TaxAmount: 840, TotalAmount: 11340, ReferencePO: "AL-2026-88",
			Items: []SalesOrderItem{
				{ItemNumber: 10, MaterialID: "MAT-1009", MaterialName: "3kVA Lithium Rackmount UPS", Quantity: 5, UOM: "EA", UnitPrice: 2100, NetAmount: 10500},
			},
			CreatedAt: baseTime.AddDate(0, 1, 1),
		},
		{
			ID: "SO-50008", CustomerID: "CUST-1009", CustomerName: "Maple Leaf Aerospace", SalesOrg: "1000",
			OrderDate: "2026-11-06", RequestedDeliveryDate: "2026-11-22", Status: "CONFIRMED", Currency: "USD",
			NetAmount: 25000, TaxAmount: 2000, TotalAmount: 27000, ReferencePO: "MLA-5510",
			Items: []SalesOrderItem{
				{ItemNumber: 10, MaterialID: "MAT-1007", MaterialName: "Sealed 42U Factory Server Cabinet", Quantity: 4, UOM: "EA", UnitPrice: 4500, NetAmount: 18000},
				{ItemNumber: 20, MaterialID: "MAT-1008", MaterialName: "Armored Fiber Optic Trunk 100m", Quantity: 10, UOM: "BOX", UnitPrice: 620, NetAmount: 6200},
				{ItemNumber: 30, MaterialID: "MAT-1015", MaterialName: "Heavy-Duty DIN Rail Mounting Kit", Quantity: 16, UOM: "BOX", UnitPrice: 50, NetAmount: 800},
			},
			CreatedAt: baseTime.AddDate(0, 1, 5),
		},
		{
			ID: "SO-50009", CustomerID: "CUST-1011", CustomerName: "Delta Port Automation", SalesOrg: "2000",
			OrderDate: "2026-11-10", RequestedDeliveryDate: "2026-11-25", Status: "OPEN", Currency: "EUR",
			NetAmount: 18000, TaxAmount: 1440, TotalAmount: 19440, ReferencePO: "DPA-TERM-09",
			Items: []SalesOrderItem{
				{ItemNumber: 10, MaterialID: "MAT-1001", MaterialName: "Industrial Edge Gateway X4", Quantity: 8, UOM: "EA", UnitPrice: 1850, NetAmount: 14800},
				{ItemNumber: 20, MaterialID: "MAT-1002", MaterialName: "Modular PLC Controller Pro", Quantity: 1, UOM: "EA", UnitPrice: 3200, NetAmount: 3200},
			},
			CreatedAt: baseTime.AddDate(0, 1, 9),
		},
		{
			ID: "SO-50010", CustomerID: "CUST-1006", CustomerName: "Nordic ColdChain AB", SalesOrg: "2000",
			OrderDate: "2026-10-08", RequestedDeliveryDate: "2026-10-20", Status: "CANCELLED", Currency: "EUR",
			NetAmount: 8900, TaxAmount: 712, TotalAmount: 9612, ReferencePO: "NCC-102", Notes: "Cancelled by customer due to site delay",
			Items: []SalesOrderItem{
				{ItemNumber: 10, MaterialID: "MAT-1012", MaterialName: "Infrared Thermal Array Sensor", Quantity: 10, UOM: "EA", UnitPrice: 890, NetAmount: 8900},
			},
			CreatedAt: baseTime.AddDate(0, 0, 7),
		},
	}
	for i := range salesOrders {
		so := salesOrders[i]
		s.salesOrders[so.ID] = &so
	}
	s.nextSOSeq = 50011

	// 5. Seed Invoices (INV-90001 .. INV-90006)
	paidTime1 := baseTime.AddDate(0, 0, 20)
	paidTime2 := baseTime.AddDate(0, 1, 2)
	invoices := []Invoice{
		{
			ID: "INV-90001", SalesOrderID: "SO-50001", CustomerID: "CUST-1001", CustomerName: "Acme Industrial Systems",
			BillingDate: "2026-10-16", DueDate: "2026-11-15", Status: "OPEN", Currency: "USD",
			NetAmount: 36000, TaxAmount: 2880, TotalAmount: 38880, PaidAmount: 0,
			Items: []InvoiceItem{
				{ItemNumber: 10, MaterialID: "MAT-1001", Description: "Industrial Edge Gateway X4", Quantity: 10, UOM: "EA", UnitPrice: 1850, NetAmount: 18500},
				{ItemNumber: 20, MaterialID: "MAT-1002", Description: "Modular PLC Controller Pro", Quantity: 5, UOM: "EA", UnitPrice: 3200, NetAmount: 16000},
				{ItemNumber: 30, MaterialID: "MAT-1003", Description: "Tri-Axial Vibration Sensor", Quantity: 6, UOM: "EA", UnitPrice: 250, NetAmount: 1500},
			},
			CreatedAt: baseTime.AddDate(0, 0, 15),
		},
		{
			ID: "INV-90002", SalesOrderID: "SO-50002", CustomerID: "CUST-1002", CustomerName: "RheinMetallurgie GmbH",
			BillingDate: "2026-10-25", DueDate: "2026-12-09", Status: "OPEN", Currency: "EUR",
			NetAmount: 48000, TaxAmount: 3840, TotalAmount: 51840, PaidAmount: 0,
			Items: []InvoiceItem{
				{ItemNumber: 10, MaterialID: "MAT-1006", Description: "SCADA Enterprise Site License", Quantity: 2, UOM: "EA", UnitPrice: 12000, NetAmount: 24000},
				{ItemNumber: 20, MaterialID: "MAT-1002", Description: "Modular PLC Controller Pro", Quantity: 6, UOM: "EA", UnitPrice: 3200, NetAmount: 19200},
				{ItemNumber: 30, MaterialID: "MAT-1003", Description: "Tri-Axial Vibration Sensor", Quantity: 20, UOM: "EA", UnitPrice: 240, NetAmount: 4800},
			},
			CreatedAt: baseTime.AddDate(0, 0, 24),
		},
		{
			ID: "INV-90003", SalesOrderID: "SO-50004", CustomerID: "CUST-1005", CustomerName: "Solstice Energy Grid",
			BillingDate: "2026-11-02", DueDate: "2027-01-01", Status: "OPEN", Currency: "USD",
			NetAmount: 85000, TaxAmount: 6800, TotalAmount: 91800, PaidAmount: 0,
			Items: []InvoiceItem{
				{ItemNumber: 10, MaterialID: "MAT-1006", Description: "SCADA Enterprise Site License", Quantity: 4, UOM: "EA", UnitPrice: 12000, NetAmount: 48000},
				{ItemNumber: 20, MaterialID: "MAT-1001", Description: "Industrial Edge Gateway X4", Quantity: 20, UOM: "EA", UnitPrice: 1850, NetAmount: 37000},
			},
			CreatedAt: baseTime.AddDate(0, 1, 1),
		},
		{
			ID: "INV-90004", SalesOrderID: "SO-50006", CustomerID: "CUST-1008", CustomerName: "Pacific Rim Semiconductors",
			BillingDate: "2026-11-12", DueDate: "2026-12-27", Status: "OPEN", Currency: "USD",
			NetAmount: 68000, TaxAmount: 5440, TotalAmount: 73440, PaidAmount: 0,
			Items: []InvoiceItem{
				{ItemNumber: 10, MaterialID: "MAT-1010", Description: "4K Machine Vision Inspection Camera", Quantity: 20, UOM: "EA", UnitPrice: 1950, NetAmount: 39000},
				{ItemNumber: 20, MaterialID: "MAT-1004", Description: "15-inch Industrial Touch Panel", Quantity: 20, UOM: "EA", UnitPrice: 1450, NetAmount: 29000},
			},
			CreatedAt: baseTime.AddDate(0, 1, 11),
		},
		{
			ID: "INV-90005", SalesOrderID: "", CustomerID: "CUST-1010", CustomerName: "Vanguard Retail Holdings",
			BillingDate: "2026-09-15", DueDate: "2026-10-15", Status: "OVERDUE", Currency: "USD",
			NetAmount: 68703.70, TaxAmount: 5496.30, TotalAmount: 74200, PaidAmount: 0,
			Items: []InvoiceItem{
				{ItemNumber: 10, MaterialID: "MAT-1001", Description: "Industrial Edge Gateway X4", Quantity: 40, UOM: "EA", UnitPrice: 1717.59, NetAmount: 68703.70},
			},
			CreatedAt: baseTime.AddDate(0, -1, 14),
		},
		{
			ID: "INV-90006", SalesOrderID: "", CustomerID: "CUST-1001", CustomerName: "Acme Industrial Systems",
			BillingDate: "2026-09-20", DueDate: "2026-10-20", Status: "PAID", Currency: "USD",
			NetAmount: 14000, TaxAmount: 1120, TotalAmount: 15120, PaidAmount: 15120,
			PaymentReference: "ACH-20261018-0041", PaidAt: &paidTime1,
			Items: []InvoiceItem{
				{ItemNumber: 10, MaterialID: "MAT-1005", Description: "24V DC 40A Redundant Power Supply", Quantity: 40, UOM: "EA", UnitPrice: 350, NetAmount: 14000},
			},
			CreatedAt: baseTime.AddDate(0, -1, 19),
		},
		{
			ID: "INV-90007", SalesOrderID: "", CustomerID: "CUST-1003", CustomerName: "Kuroda Precision Robotics",
			BillingDate: "2026-10-01", DueDate: "2026-10-31", Status: "PAID", Currency: "USD",
			NetAmount: 28000, TaxAmount: 2240, TotalAmount: 30240, PaidAmount: 30240,
			PaymentReference: "WIRE-JP-99201", PaidAt: &paidTime2,
			Items: []InvoiceItem{
				{ItemNumber: 10, MaterialID: "MAT-1011", Description: "5kW AC Servo Drive Unit", Quantity: 16, UOM: "EA", UnitPrice: 1750, NetAmount: 28000},
			},
			CreatedAt: baseTime,
		},
	}
	for i := range invoices {
		inv := invoices[i]
		s.invoices[inv.ID] = &inv
	}
	s.nextInvoiceSeq = 90008

	// 6. Seed Purchase Orders (PO-45001 .. PO-45006)
	purchaseOrders := []PurchaseOrder{
		{
			ID: "PO-45001", VendorID: "VEND-1001", VendorName: "SiliconFab Precision Corp", PurchasingOrg: "1000", Plant: "2000",
			OrderDate: "2026-11-05", ExpectedDeliveryDate: "2026-11-22", Status: "ORDERED", Currency: "USD", TotalAmount: 37600,
			Notes: "Replenishment for low-stock HMI panels and PCB substrates",
			Items: []PurchaseOrderItem{
				{ItemNumber: 10, MaterialID: "MAT-1004", MaterialName: "15-inch Industrial Touch Panel", Quantity: 30, UOM: "EA", UnitPrice: 940, NetAmount: 28200},
				{ItemNumber: 20, MaterialID: "MAT-1014", MaterialName: "8-Layer FR4 Bare PCB Panel", Quantity: 100, UOM: "SET", UnitPrice: 94, NetAmount: 9400},
			},
			CreatedAt: baseTime.AddDate(0, 1, 4),
		},
		{
			ID: "PO-45002", VendorID: "VEND-1002", VendorName: "Bavaria Stahl & Alu GmbH", PurchasingOrg: "1000", Plant: "2000",
			OrderDate: "2026-11-08", ExpectedDeliveryDate: "2026-11-26", Status: "ORDERED", Currency: "EUR", TotalAmount: 58500,
			Notes: "Factory server cabinets restock",
			Items: []PurchaseOrderItem{
				{ItemNumber: 10, MaterialID: "MAT-1007", MaterialName: "Sealed 42U Factory Server Cabinet", Quantity: 15, UOM: "EA", UnitPrice: 2900, NetAmount: 43500},
				{ItemNumber: 20, MaterialID: "MAT-1013", MaterialName: "6061-T6 Extruded Aluminum Profile", Quantity: 250, UOM: "EA", UnitPrice: 60, NetAmount: 15000},
			},
			CreatedAt: baseTime.AddDate(0, 1, 7),
		},
		{
			ID: "PO-45003", VendorID: "VEND-1003", VendorName: "OptiLens Photonics Ltd", PurchasingOrg: "1000", Plant: "1000",
			OrderDate: "2026-11-12", ExpectedDeliveryDate: "2026-11-28", Status: "APPROVED", Currency: "USD", TotalAmount: 14500,
			Notes: "Thermal array sensors urgent reorder",
			Items: []PurchaseOrderItem{
				{ItemNumber: 10, MaterialID: "MAT-1012", MaterialName: "Infrared Thermal Array Sensor", Quantity: 25, UOM: "EA", UnitPrice: 580, NetAmount: 14500},
			},
			CreatedAt: baseTime.AddDate(0, 1, 11),
		},
		{
			ID: "PO-45004", VendorID: "VEND-1004", VendorName: "Midwest Power Components", PurchasingOrg: "1000", Plant: "1000",
			OrderDate: "2026-10-12", ExpectedDeliveryDate: "2026-10-22", Status: "RECEIVED", Currency: "USD", TotalAmount: 24500,
			Items: []PurchaseOrderItem{
				{ItemNumber: 10, MaterialID: "MAT-1005", MaterialName: "24V DC 40A Redundant Power Supply", Quantity: 100, UOM: "EA", UnitPrice: 245, NetAmount: 24500},
			},
			CreatedAt: baseTime.AddDate(0, 0, 11),
		},
		{
			ID: "PO-45005", VendorID: "VEND-1006", VendorName: "Shenzhen FiberLink Co", PurchasingOrg: "1000", Plant: "1000",
			OrderDate: "2026-11-15", ExpectedDeliveryDate: "2026-12-05", Status: "DRAFT", Currency: "USD", TotalAmount: 19000,
			Items: []PurchaseOrderItem{
				{ItemNumber: 10, MaterialID: "MAT-1008", MaterialName: "Armored Fiber Optic Trunk 100m", Quantity: 50, UOM: "BOX", UnitPrice: 380, NetAmount: 19000},
			},
			CreatedAt: baseTime.AddDate(0, 1, 14),
		},
	}
	for i := range purchaseOrders {
		po := purchaseOrders[i]
		s.purchaseOrders[po.ID] = &po
	}
	s.nextPOSeq = 45006

	s.loadExtraSeedDir()
}

func loadJSONSlice[T any](path string, dst map[string]*T, idOf func(*T) string) {
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var items []T
	if err := json.Unmarshal(b, &items); err != nil {
		log.Printf("warning: skipping %s: %v", path, err)
		return
	}
	for i := range items {
		item := items[i]
		id := idOf(&item)
		if id != "" {
			dst[id] = &item
		}
	}
}

// loadExtraSeedDir optionally loads additional JSON seed files from SEED_DIR (default "./data").
func (s *Store) loadExtraSeedDir() {
	dir := os.Getenv("SEED_DIR")
	if dir == "" {
		dir = "data"
	}
	if _, err := os.Stat(dir); err != nil {
		return
	}
	loadJSONSlice(filepath.Join(dir, "customers.json"), s.customers, func(c *Customer) string { return c.ID })
	loadJSONSlice(filepath.Join(dir, "materials.json"), s.materials, func(m *Material) string { return m.ID })
	loadJSONSlice(filepath.Join(dir, "vendors.json"), s.vendors, func(v *Vendor) string { return v.ID })
	loadJSONSlice(filepath.Join(dir, "sales_orders.json"), s.salesOrders, func(so *SalesOrder) string { return so.ID })
	loadJSONSlice(filepath.Join(dir, "invoices.json"), s.invoices, func(inv *Invoice) string { return inv.ID })
	loadJSONSlice(filepath.Join(dir, "purchase_orders.json"), s.purchaseOrders, func(po *PurchaseOrder) string { return po.ID })
}
