-- Add test events for risk objects to populate charts and history

-- First, ensure we have some detections
INSERT OR IGNORE INTO detections (id, name, description, query, status, severity, risk_points, risk_object) 
VALUES 
    (101, 'Suspicious Login Activity', 'Multiple failed login attempts detected', 'event.type="login" AND event.outcome="failure" AND event.count > 5', 'production', 'high', 10, 'User'),
    (102, 'Data Exfiltration Attempt', 'Large data transfer to external IP', 'network.bytes_out > 1000000 AND destination.ip NOT IN internal_ranges', 'production', 'critical', 15, 'Host'),
    (103, 'Privilege Escalation', 'User privilege changes detected', 'event.action="privilege_assigned" AND user.privilege_level="admin"', 'production', 'high', 12, 'User');

-- Add events for different risk objects with varying timestamps
-- Events for User: admin (id=5, current_score=60)
INSERT INTO events (detection_id, entity_id, timestamp, risk_points, raw_data, is_false_positive)
VALUES 
    (101, 5, datetime('now', '-7 days'), 10, '{"user":"admin","action":"failed_login","count":6}', 0),
    (101, 5, datetime('now', '-6 days'), 10, '{"user":"admin","action":"failed_login","count":8}', 0),
    (102, 5, datetime('now', '-5 days'), 15, '{"user":"admin","action":"data_transfer","bytes":2000000}', 0),
    (101, 5, datetime('now', '-4 days'), 10, '{"user":"admin","action":"failed_login","count":7}', 0),
    (103, 5, datetime('now', '-3 days'), 15, '{"user":"admin","action":"privilege_escalation","new_role":"domain_admin"}', 0),
    (101, 5, datetime('now', '-2 days'), 5, '{"user":"admin","action":"failed_login","count":5}', 1), -- False positive
    (102, 5, datetime('now', '-1 days'), 15, '{"user":"admin","action":"data_transfer","bytes":5000000}', 0),
    (101, 5, datetime('now', '-12 hours'), 10, '{"user":"admin","action":"suspicious_login","source":"external_ip"}', 0);

-- Events for Host: SERVER-DC01 (id=2, current_score=45)
INSERT INTO events (detection_id, entity_id, timestamp, risk_points, raw_data, is_false_positive)
VALUES 
    (101, 2, datetime('now', '-10 days'), 8, '{"host":"SERVER-DC01","action":"authentication_failure","service":"RDP"}', 0),
    (102, 2, datetime('now', '-8 days'), 12, '{"host":"SERVER-DC01","action":"unusual_network_activity","connections":150}', 0),
    (103, 2, datetime('now', '-6 days'), 10, '{"host":"SERVER-DC01","action":"service_creation","service":"suspicious_svc"}', 0),
    (101, 2, datetime('now', '-4 days'), 8, '{"host":"SERVER-DC01","action":"authentication_failure","service":"SMB"}', 0),
    (102, 2, datetime('now', '-2 days'), 7, '{"host":"SERVER-DC01","action":"port_scan_detected","ports_scanned":100}', 0);

-- Events for IP: 192.168.1.100 (id=7, current_score=30)
INSERT INTO events (detection_id, entity_id, timestamp, risk_points, raw_data, is_false_positive)
VALUES 
    (102, 7, datetime('now', '-5 days'), 10, '{"ip":"192.168.1.100","action":"suspicious_traffic","protocol":"uncommon"}', 0),
    (101, 7, datetime('now', '-3 days'), 8, '{"ip":"192.168.1.100","action":"brute_force_attempt","target_ports":[22,3389]}', 0),
    (102, 7, datetime('now', '-1 days'), 12, '{"ip":"192.168.1.100","action":"data_exfiltration","volume":"high"}', 0);

-- Events for Host: WORKSTATION-01 (id=1, current_score=25)
INSERT INTO events (detection_id, entity_id, timestamp, risk_points, raw_data, is_false_positive)
VALUES 
    (101, 1, datetime('now', '-14 days'), 5, '{"host":"WORKSTATION-01","action":"login_anomaly","time":"02:00"}', 0),
    (103, 1, datetime('now', '-10 days'), 8, '{"host":"WORKSTATION-01","action":"registry_modification","key":"Run"}', 0),
    (101, 1, datetime('now', '-7 days'), 5, '{"host":"WORKSTATION-01","action":"failed_auth","attempts":4}', 0),
    (102, 1, datetime('now', '-3 days'), 7, '{"host":"WORKSTATION-01","action":"unusual_dns_query","domain":"suspicious.com"}', 0);

-- Events for User: john.doe (id=4, current_score=20)
INSERT INTO events (detection_id, entity_id, timestamp, risk_points, raw_data, is_false_positive)
VALUES 
    (101, 4, datetime('now', '-9 days'), 5, '{"user":"john.doe","action":"password_spray_victim","source":"external"}', 0),
    (102, 4, datetime('now', '-6 days'), 8, '{"user":"john.doe","action":"abnormal_file_access","files_accessed":200}', 0),
    (101, 4, datetime('now', '-2 days'), 7, '{"user":"john.doe","action":"impossible_travel","locations":["NYC","London"]}', 0);

-- Update the last_seen timestamps for risk objects
UPDATE risk_objects SET last_seen = datetime('now') WHERE id IN (1,2,4,5,7);

-- Add some risk alerts for high-scoring entities
INSERT INTO risk_alerts (entity_id, triggered_at, total_score, status)
VALUES 
    (5, datetime('now', '-2 days'), 65, 'Investigation'), -- admin user
    (2, datetime('now', '-4 days'), 48, 'Triage'); -- SERVER-DC01

SELECT 'Test events added successfully!' as message;
SELECT COUNT(*) as total_events_added FROM events WHERE datetime(timestamp) > datetime('now', '-15 days');