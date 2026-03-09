# Industry Quick Reference

Summary of tables, key columns, and typical sizes for each supported industry. For complete column-level schemas with generation hints and constraints, see the [industry-patterns/](industry-patterns/) directory and `assets/schemas/` JSON files.

## Retail Tables

| Table          | Key Columns                                        | Typical Size |
| -------------- | -------------------------------------------------- | ------------ |
| `customers`    | customer_id, name, email, loyalty_tier, region     | 100K-1M      |
| `transactions` | txn_id, customer_id, product_id, amount, timestamp | 1M-100M      |
| `products`     | product_id, name, category, price, sku             | 10K-100K     |
| `inventory`    | product_id, store_id, quantity, last_updated       | 100K-1M      |

## Healthcare Tables

| Table        | Key Columns                                 | Notes                |
| ------------ | ------------------------------------------- | -------------------- |
| `patients`   | patient_id, name, dob, mrn                  | HIPAA-safe synthetic |
| `encounters` | encounter_id, patient_id, provider_id, date | Clinical visits      |
| `claims`     | claim_id, encounter_id, icd10_code, amount  | Insurance claims     |

## Financial Tables

| Table          | Key Columns                                   | Use Case          |
| -------------- | --------------------------------------------- | ----------------- |
| `accounts`     | account_id, customer_id, type, balance        | Customer accounts |
| `trades`       | trade_id, account_id, symbol, quantity, price | Trading activity  |
| `transactions` | txn_id, account_id, type, amount, timestamp   | Money movement    |

## IoT Tables

| Table             | Key Columns                                    | Pattern              |
| ----------------- | ---------------------------------------------- | -------------------- |
| `devices`         | device_id, device_type, location, install_date | Device registry      |
| `sensor_readings` | device_id, timestamp, metric, value            | Time-series data     |
| `events`          | event_id, device_id, event_type, severity      | Alerts/notifications |
| `telemetry`       | device_id, timestamp, lat, lon, speed          | GPS/telematics       |

## Manufacturing Tables

| Table                 | Key Columns                                          | Pattern                           |
| --------------------- | ---------------------------------------------------- | --------------------------------- |
| `equipment`           | equipment_id, type, manufacturer, zone, install_date | Asset registry                    |
| `sensor_data`         | equipment_id, timestamp, sensors A-F, is_anomaly     | Multi-sensor with fault injection |
| `maintenance_records` | work_order_id, equipment_id, type, priority, cost    | Predictive maintenance            |

## Supply Chain (CPG) Tables

| Table                 | Key Columns                                                    | Typical Size | Pattern                     |
| --------------------- | -------------------------------------------------------------- | ------------ | --------------------------- |
| `products`            | id, sku, category, brand, unit_cost, unit_price                | 500          | CPG product master          |
| `distribution_centers`| id, dc_code, region, capacity_pallets, utilization_pct         | 25           | DC network                  |
| `stores`              | id, store_code, format, distribution_center_id                 | 1K           | Retail locations            |
| `orders`              | id, dc_id, product_id, scheduled_date, actual_date, status     | 10K          | Manufacturing/purchase orders |
| `inventory_snapshots` | id, product_id, dc_id, store_id, quantity, stockout_risk       | 50K          | Point-in-time inventory     |
| `shipments`           | id, dc_id, store_id, product_id, transport_mode, carrier       | 30K          | Logistics & delivery tracking |

## Oil & Gas Tables

| Table              | Key Columns                                                | Pattern                          |
| ------------------ | ---------------------------------------------------------- | -------------------------------- |
| `well_headers`     | well_id, api_number, operator, formation, spud_date        | Well metadata (per-formation)    |
| `daily_production` | well_id, production_date, oil_bbl, gas_mcf, water_bbl     | ARPS decline curve production    |
| `type_curves`      | formation, month_on_production, oil_rate, gas_rate         | Forecast by formation            |

## Gaming Tables

| Table          | Key Columns                                                     | Pattern                           |
| -------------- | --------------------------------------------------------------- | --------------------------------- |
| `login_events` | account_id, device_id, ip_address, country, login_ts, platform  | Hash-based IDs, country-weighted (Tier 3 only) |

## Clinical Trials Tables (Healthcare Variant)

| Table                | Key Columns                                                  | Pattern                      |
| -------------------- | ------------------------------------------------------------ | ---------------------------- |
| `clinical_trials`    | trial_id, phase, therapeutic_area, sponsor, status           | Trial protocol metadata      |
| `study_sites`        | site_id, trial_id, pi_name, institution, enrollment_capacity | Investigator site network    |
| `study_participants` | participant_id, site_id, arm, consent_date, severity         | Randomized participants      |
| `adverse_events`     | ae_id, participant_id, term, severity, outcome, onset_date   | Correlated AE patterns       |
| `lab_measurements`   | lab_id, participant_id, test_name, value, unit, ref_range    | Treatment-effect lab results |
