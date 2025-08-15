-- Sample data for DetectionMatrix

-- Insert sample MITRE ATT&CK techniques
INSERT OR IGNORE INTO mitre_techniques (id, name, description, tactic, domain) VALUES
('T1059.001', 'PowerShell', 'Adversaries may abuse PowerShell commands and scripts for execution.', 'Execution', 'Enterprise'),
('T1059.003', 'Windows Command Shell', 'Adversaries may abuse the Windows command shell for execution.', 'Execution', 'Enterprise'),
('T1003.001', 'LSASS Memory', 'Adversaries may attempt to access credential material stored in LSASS memory.', 'Credential Access', 'Enterprise'),
('T1055', 'Process Injection', 'Adversaries may inject code into processes to evade defenses.', 'Defense Evasion', 'Enterprise'),
('T1082', 'System Information Discovery', 'An adversary may attempt to get detailed information about the operating system.', 'Discovery', 'Enterprise'),
('T1190', 'Exploit Public-Facing Application', 'Adversaries may attempt to exploit a weakness in an Internet-facing host.', 'Initial Access', 'Enterprise'),
('T1566.001', 'Spearphishing Attachment', 'Adversaries may send spearphishing emails with a malicious attachment.', 'Initial Access', 'Enterprise'),
('T1070.004', 'File Deletion', 'Adversaries may delete files left behind by the actions of their intrusion activity.', 'Defense Evasion', 'Enterprise'),
('T1033', 'System Owner/User Discovery', 'Adversaries may attempt to identify the primary user of a system.', 'Discovery', 'Enterprise'),
('T1087.002', 'Domain Account', 'Adversaries may attempt to get a listing of domain accounts.', 'Discovery', 'Enterprise');

-- Insert sample data sources
INSERT OR IGNORE INTO data_sources (name, description, log_format) VALUES
('Windows Event Logs', 'Windows Security, System, and Application event logs', 'EVTX'),
('Sysmon', 'System Monitor (Sysmon) for Windows', 'EVTX'),
('PowerShell Logs', 'PowerShell script block and module logging', 'EVTX'),
('Process Creation', 'Process creation events from EDR/Sysmon', 'JSON'),
('Network Traffic', 'Network connection and traffic logs', 'JSON'),
('File System', 'File creation, modification, and deletion events', 'JSON'),
('Registry', 'Windows Registry modification events', 'JSON'),
('Authentication', 'Login and authentication events', 'JSON');

-- Insert sample detections
INSERT OR IGNORE INTO detections (name, description, query, status, severity, risk_points, playbook_link, owner, risk_object) VALUES
('Suspicious PowerShell Execution', 'Detects potentially malicious PowerShell commands', 'EventCode=4688 AND CommandLine CONTAINS "powershell" AND (CommandLine CONTAINS "-enc" OR CommandLine CONTAINS "IEX" OR CommandLine CONTAINS "Invoke-Expression")', 'production', 'High', 15, 'https://playbooks.example.com/powershell', 'SOC Team', 'Host'),
('Credential Dumping - LSASS Access', 'Detects attempts to access LSASS memory for credential dumping', 'EventCode=4656 AND ObjectName CONTAINS "lsass.exe" AND AccessMask=0x1010', 'production', 'Critical', 25, 'https://playbooks.example.com/lsass', 'SOC Team', 'Host'),
('Process Injection Detection', 'Identifies potential process injection techniques', 'EventCode=8 AND (TargetImage CONTAINS "explorer.exe" OR TargetImage CONTAINS "winlogon.exe") AND SourceImage NOT CONTAINS "System"', 'test', 'High', 20, 'https://playbooks.example.com/injection', 'Detection Team', 'Host'),
('Suspicious Network Connections', 'Detects connections to suspicious IP addresses', 'EventCode=3 AND (DestinationIp STARTSWITH "192.168" OR DestinationIp STARTSWITH "10.") AND NOT (SourceImage CONTAINS "chrome.exe" OR SourceImage CONTAINS "firefox.exe")', 'production', 'Medium', 10, 'https://playbooks.example.com/network', 'SOC Team', 'IP'),
('Failed Login Attempts', 'Multiple failed login attempts from single source', 'EventCode=4625 AND count() > 5 GROUP BY SourceIP WINDOW 5m', 'production', 'Medium', 8, 'https://playbooks.example.com/bruteforce', 'SOC Team', 'IP'),
('File Deletion After Execution', 'Detects file deletion shortly after execution', 'EventCode=2 AND EventCode=23 WITHIN 30s WHERE Image=TargetFilename', 'draft', 'Low', 5, '', 'Detection Team', 'Host'),
('Encoded Command Execution', 'Detects base64 encoded command execution', 'CommandLine CONTAINS "-EncodedCommand" OR CommandLine CONTAINS "-enc"', 'test', 'High', 18, 'https://playbooks.example.com/encoded', 'Detection Team', 'Host'),
('Admin Account Enumeration', 'Detects enumeration of administrative accounts', 'EventCode=4798 OR EventCode=4799 AND count() > 3 GROUP BY SubjectUserName', 'production', 'Medium', 12, 'https://playbooks.example.com/enum', 'SOC Team', 'User');

-- Map detections to MITRE techniques
INSERT OR IGNORE INTO detection_mitre_map (detection_id, mitre_id) VALUES
(1, 'T1059.001'), -- PowerShell detection -> PowerShell technique
(2, 'T1003.001'), -- LSASS detection -> LSASS Memory technique
(3, 'T1055'),     -- Process injection -> Process Injection technique
(4, 'T1082'),     -- Network connections -> System Info Discovery
(5, 'T1190'),     -- Failed logins -> Exploit Public-Facing Application
(6, 'T1070.004'), -- File deletion -> File Deletion technique
(7, 'T1059.001'), -- Encoded commands -> PowerShell technique
(8, 'T1087.002'); -- Account enum -> Domain Account technique

-- Map detections to data sources
INSERT OR IGNORE INTO detection_datasource (detection_id, datasource_id) VALUES
(1, 1), (1, 2), -- PowerShell detection uses Windows Events and Sysmon
(2, 1), (2, 2), -- LSASS detection uses Windows Events and Sysmon
(3, 2),         -- Process injection uses Sysmon
(4, 5),         -- Network detection uses Network Traffic
(5, 8),         -- Failed logins use Authentication logs
(6, 6),         -- File deletion uses File System logs
(7, 1), (7, 3), -- Encoded commands use Windows Events and PowerShell logs
(8, 1);         -- Account enum uses Windows Events

-- Insert sample risk objects
INSERT OR IGNORE INTO risk_objects (entity_type, entity_value, current_score) VALUES
('Host', 'WORKSTATION-01', 25),
('Host', 'SERVER-DC01', 45),
('Host', 'LAPTOP-USER01', 15),
('User', 'john.doe', 20),
('User', 'admin', 60),
('User', 'service.account', 10),
('IP', '192.168.1.100', 30),
('IP', '10.0.0.50', 12),
('IP', '172.16.1.200', 8);

-- Insert sample events
INSERT OR IGNORE INTO events (detection_id, entity_id, risk_points, raw_data, context) VALUES
(1, 1, 15, '{"commandline": "powershell.exe -enc SGVsbG8gV29ybGQ="}', '{"process": "powershell.exe", "user": "john.doe"}'),
(2, 2, 25, '{"process": "mimikatz.exe", "target": "lsass.exe"}', '{"severity": "critical", "user": "admin"}'),
(4, 7, 10, '{"source_ip": "192.168.1.100", "dest_ip": "8.8.8.8"}', '{"protocol": "TCP", "port": 443}'),
(5, 8, 8, '{"failed_attempts": 6, "username": "admin"}', '{"source_ip": "10.0.0.50", "timespan": "5min"}'),
(1, 3, 15, '{"commandline": "powershell.exe IEX (New-Object Net.WebClient).DownloadString(...)"}', '{"process": "powershell.exe", "user": "user01"}');

-- Insert sample risk alerts
INSERT OR IGNORE INTO risk_alerts (entity_id, total_score, status, notes, owner) VALUES
(2, 45, 'Investigation', 'Multiple suspicious activities detected on domain controller', 'analyst.smith'),
(5, 60, 'Incident', 'Admin account showing signs of compromise', 'incident.responder'),
(7, 30, 'Triage', 'Unusual network traffic patterns detected', 'soc.analyst');

-- Update detection event counts
UPDATE detections SET 
    event_count_last_30_days = (
        SELECT COUNT(*) FROM events 
        WHERE detection_id = detections.id 
        AND timestamp > datetime('now', '-30 days')
    );

-- Insert sample false positives
INSERT OR IGNORE INTO false_positives (event_id, reason, analyst_name) VALUES
(1, 'Legitimate admin script execution', 'analyst.jones'),
(3, 'Expected network traffic for application updates', 'soc.analyst');