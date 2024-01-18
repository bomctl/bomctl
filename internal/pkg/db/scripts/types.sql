-- ------------------------------------------------------------------------
-- SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
-- SPDX-FileName: internal/pkg/db/scripts/types.sql
-- SPDX-FileType: SOURCE
-- SPDX-License-Identifier: Apache-2.0
-- ------------------------------------------------------------------------
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
-- http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.
-- ------------------------------------------------------------------------
PRAGMA foreign_keys = ON;

-- @block create `sbom_type` table
-- @conn bomctl
-- @label Initialize `sbom_type` table
-- #----------------------------------------------#
-- #----[ SBOMType enum ]-------------------------#
-- #----------------------------------------------#
-- @label Create table
CREATE TABLE IF NOT EXISTS sbom_type (id INTEGER PRIMARY KEY, name TEXT);

-- @label Populate values
INSERT OR IGNORE INTO sbom_type (id, name) VALUES(0, 'OTHER');
INSERT OR IGNORE INTO sbom_type (id, name) VALUES(1, 'DESIGN');
INSERT OR IGNORE INTO sbom_type (id, name) VALUES(2, 'SOURCE');
INSERT OR IGNORE INTO sbom_type (id, name) VALUES(3, 'BUILD');
INSERT OR IGNORE INTO sbom_type (id, name) VALUES(4, 'ANALYZED');
INSERT OR IGNORE INTO sbom_type (id, name) VALUES(5, 'DEPLOYED');
INSERT OR IGNORE INTO sbom_type (id, name) VALUES(6, 'RUNTIME');
INSERT OR IGNORE INTO sbom_type (id, name) VALUES(7, 'DISCOVERY');
INSERT OR IGNORE INTO sbom_type (id, name) VALUES(8, 'DECOMISSION');

-- @label Define triggers
CREATE TRIGGER IF NOT EXISTS tr_sbom_type_before_update
BEFORE UPDATE ON sbom_type
BEGIN
    SELECT RAISE(IGNORE);
END;

CREATE TRIGGER IF NOT EXISTS tr_sbom_type_before_insert
BEFORE INSERT ON sbom_type
BEGIN
    SELECT RAISE(IGNORE);
END;

-- @block create `external_reference_type` table
-- @conn bomctl
-- @label Initialize `external_reference_type` table
-- #----------------------------------------------#
-- #----[ ExternalReferenceType enum ]------------#
-- #----------------------------------------------#
-- @label Create table
CREATE TABLE IF NOT EXISTS external_reference_type (
    id   INTEGER PRIMARY KEY,
    name TEXT
);

-- @label Populate values
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (0, 'UNKNOWN');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (1, 'ATTESTATION');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (2, 'BINARY');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (3, 'BOM');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (4, 'BOWER');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (5, 'BUILD_META');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (6, 'BUILD_SYSTEM');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (7, 'CERTIFICATION_REPORT');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (8, 'CHAT');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (9, 'CODIFIED_INFRASTRUCTURE');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (10, 'COMPONENT_ANALYSIS_REPORT');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (11, 'CONFIGURATION');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (12, 'DISTRIBUTION_INTAKE');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (13, 'DOCUMENTATION');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (14, 'DOWNLOAD');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (15, 'DYNAMIC_ANALYSIS_REPORT');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (16, 'EOL_NOTICE');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (17, 'EVIDENCE');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (18, 'EXPORT_CONTROL_ASSESSMENT');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (19, 'FORMULATION');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (20, 'FUNDING');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (21, 'ISSUE_TRACKER');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (22, 'LICENSE');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (23, 'LOG');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (24, 'MAILING_LIST');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (25, 'MATURITY_REPORT');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (26, 'MAVEN_CENTRAL');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (27, 'METRICS');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (28, 'MODEL_CARD');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (29, 'NPM');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (30, 'NUGET');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (31, 'OTHER');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (32, 'POAM');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (33, 'PRIVACY_ASSESSMENT');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (34, 'PRODUCT_METADATA');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (35, 'PURCHASE_ORDER');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (36, 'QUALITY_ASSESSMENT_REPORT');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (37, 'QUALITY_METRICS');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (38, 'RELEASE_HISTORY');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (39, 'RELEASE_NOTES');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (40, 'RISK_ASSESSMENT');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (41, 'RUNTIME_ANALYSIS_REPORT');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (42, 'SECURE_SOFTWARE_ATTESTATION');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (43, 'SECURITY_ADVERSARY_MODEL');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (44, 'SECURITY_ADVISORY');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (45, 'SECURITY_CONTACT');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (46, 'SECURITY_FIX');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (47, 'SECURITY_OTHER');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (48, 'SECURITY_PENTEST_REPORT');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (49, 'SECURITY_POLICY');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (50, 'SECURITY_SWID');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (51, 'SECURITY_THREAT_MODEL');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (52, 'SOCIAL');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (53, 'SOURCE_ARTIFACT');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (54, 'STATIC_ANALYSIS_REPORT');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (55, 'SUPPORT');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (56, 'VCS');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (57, 'VULNERABILITY_ASSERTION');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (58, 'VULNERABILITY_DISCLOSURE_REPORT');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (59, 'VULNERABILITY_EXPLOITABILITY_ASSESSMENT');
INSERT OR IGNORE INTO external_reference_type (id, name) VALUES (60, 'WEBSITE');

-- @label Define triggers
CREATE TRIGGER IF NOT EXISTS tr_external_reference_type_before_update
BEFORE UPDATE ON external_reference_type
BEGIN
    SELECT RAISE(IGNORE);
END;

CREATE TRIGGER IF NOT EXISTS tr_external_reference_type_before_insert
BEFORE INSERT ON external_reference_type
BEGIN
    SELECT RAISE(IGNORE);
END;

-- @block create `hash_algorithm` table
-- @conn bomctl
-- @label Initialize `hash_algorithm` table
-- #----------------------------------------------#
-- #----[ HashAlgorithm enum ]--------------------#
-- #----------------------------------------------#
-- @label Create table
CREATE TABLE IF NOT EXISTS hash_algorithm (
    id   INTEGER PRIMARY KEY,
    name TEXT
);

-- @label Populate values
INSERT OR IGNORE INTO hash_algorithm (id, name) VALUES (0,  'UNKNOWN');
INSERT OR IGNORE INTO hash_algorithm (id, name) VALUES (1,  'MD5');
INSERT OR IGNORE INTO hash_algorithm (id, name) VALUES (2,  'SHA1');
INSERT OR IGNORE INTO hash_algorithm (id, name) VALUES (3,  'SHA256');
INSERT OR IGNORE INTO hash_algorithm (id, name) VALUES (4,  'SHA384');
INSERT OR IGNORE INTO hash_algorithm (id, name) VALUES (5,  'SHA512');
INSERT OR IGNORE INTO hash_algorithm (id, name) VALUES (6,  'SHA3_256');
INSERT OR IGNORE INTO hash_algorithm (id, name) VALUES (7,  'SHA3_384');
INSERT OR IGNORE INTO hash_algorithm (id, name) VALUES (8,  'SHA3_512');
INSERT OR IGNORE INTO hash_algorithm (id, name) VALUES (9,  'BLAKE2B_256');
INSERT OR IGNORE INTO hash_algorithm (id, name) VALUES (10, 'BLAKE2B_384');
INSERT OR IGNORE INTO hash_algorithm (id, name) VALUES (11, 'BLAKE2B_512');
INSERT OR IGNORE INTO hash_algorithm (id, name) VALUES (12, 'BLAKE3');
INSERT OR IGNORE INTO hash_algorithm (id, name) VALUES (13, 'MD2');
INSERT OR IGNORE INTO hash_algorithm (id, name) VALUES (14, 'ADLER32');
INSERT OR IGNORE INTO hash_algorithm (id, name) VALUES (15, 'MD4');
INSERT OR IGNORE INTO hash_algorithm (id, name) VALUES (16, 'MD6');
INSERT OR IGNORE INTO hash_algorithm (id, name) VALUES (17, 'SHA224');

-- @label Define triggers
CREATE TRIGGER IF NOT EXISTS tr_hash_algorithm_before_update
BEFORE UPDATE ON hash_algorithm
BEGIN
    SELECT RAISE(IGNORE);
END;

CREATE TRIGGER IF NOT EXISTS tr_hash_algorithm_before_insert
BEFORE INSERT ON hash_algorithm
BEGIN
    SELECT RAISE(IGNORE);
END;

-- @block create `software_identifier_type` table
-- @conn bomctl
-- @label Initialize `software_identifier_type` table
-- #----------------------------------------------#
-- #----[ SoftwareIdentifierType enum ]-----------#
-- #----------------------------------------------#
-- @label Create table
CREATE TABLE IF NOT EXISTS software_identifier_type (
    id   INTEGER PRIMARY KEY,
    name TEXT
);

-- @label Populate values
INSERT OR IGNORE INTO software_identifier_type (id, name) VALUES (0, 'UNKNOWN_IDENTIFIER_TYPE');
INSERT OR IGNORE INTO software_identifier_type (id, name) VALUES (1, 'PURL');
INSERT OR IGNORE INTO software_identifier_type (id, name) VALUES (2, 'CPE22');
INSERT OR IGNORE INTO software_identifier_type (id, name) VALUES (3, 'CPE23');
INSERT OR IGNORE INTO software_identifier_type (id, name) VALUES (4, 'GITOID');

-- @label Define triggers
CREATE TRIGGER IF NOT EXISTS tr_software_identifier_type_before_update
BEFORE UPDATE ON software_identifier_type
BEGIN
    SELECT RAISE(IGNORE);
END;

CREATE TRIGGER IF NOT EXISTS tr_software_identifier_type_before_insert
BEFORE INSERT ON software_identifier_type
BEGIN
    SELECT RAISE(IGNORE);
END;

-- @block create `purpose` table
-- @conn bomctl
-- @label Initialize `purpose` table
-- #----------------------------------------------#
-- #----[ Purpose enum ]--------------------------#
-- #----------------------------------------------#
-- @label Create table
CREATE TABLE IF NOT EXISTS purpose (
    id   INTEGER PRIMARY KEY,
    name TEXT
);

-- @label Populate values
INSERT OR IGNORE INTO purpose (id, name) VALUES (0,  'UNKNOWN_PURPOSE');
INSERT OR IGNORE INTO purpose (id, name) VALUES (1,  'APPLICATION');
INSERT OR IGNORE INTO purpose (id, name) VALUES (2,  'ARCHIVE');
INSERT OR IGNORE INTO purpose (id, name) VALUES (3,  'BOM');
INSERT OR IGNORE INTO purpose (id, name) VALUES (4,  'CONFIGURATION');
INSERT OR IGNORE INTO purpose (id, name) VALUES (5,  'CONTAINER');
INSERT OR IGNORE INTO purpose (id, name) VALUES (6,  'DATA');
INSERT OR IGNORE INTO purpose (id, name) VALUES (7,  'DEVICE');
INSERT OR IGNORE INTO purpose (id, name) VALUES (8,  'DEVICE_DRIVER');
INSERT OR IGNORE INTO purpose (id, name) VALUES (9,  'DOCUMENTATION');
INSERT OR IGNORE INTO purpose (id, name) VALUES (10, 'EVIDENCE');
INSERT OR IGNORE INTO purpose (id, name) VALUES (11, 'EXECUTABLE');
INSERT OR IGNORE INTO purpose (id, name) VALUES (12, 'FILE');
INSERT OR IGNORE INTO purpose (id, name) VALUES (13, 'FIRMWARE');
INSERT OR IGNORE INTO purpose (id, name) VALUES (14, 'FRAMEWORK');
INSERT OR IGNORE INTO purpose (id, name) VALUES (15, 'INSTALL');
INSERT OR IGNORE INTO purpose (id, name) VALUES (16, 'LIBRARY');
INSERT OR IGNORE INTO purpose (id, name) VALUES (17, 'MACHINE_LEARNING_MODEL');
INSERT OR IGNORE INTO purpose (id, name) VALUES (18, 'MANIFEST');
INSERT OR IGNORE INTO purpose (id, name) VALUES (19, 'MODEL');
INSERT OR IGNORE INTO purpose (id, name) VALUES (20, 'MODULE');
INSERT OR IGNORE INTO purpose (id, name) VALUES (21, 'OPERATING_SYSTEM');
INSERT OR IGNORE INTO purpose (id, name) VALUES (22, 'OTHER');
INSERT OR IGNORE INTO purpose (id, name) VALUES (23, 'PATCH');
INSERT OR IGNORE INTO purpose (id, name) VALUES (24, 'PLATFORM');
INSERT OR IGNORE INTO purpose (id, name) VALUES (25, 'REQUIREMENT');
INSERT OR IGNORE INTO purpose (id, name) VALUES (26, 'SOURCE');
INSERT OR IGNORE INTO purpose (id, name) VALUES (27, 'SPECIFICATION');
INSERT OR IGNORE INTO purpose (id, name) VALUES (28, 'TEST');

-- @label Define triggers
CREATE TRIGGER IF NOT EXISTS tr_purpose_before_update
BEFORE UPDATE ON purpose
BEGIN
    SELECT RAISE(IGNORE);
END;

CREATE TRIGGER IF NOT EXISTS tr_purpose_before_insert
BEFORE INSERT ON purpose
BEGIN
    SELECT RAISE(IGNORE);
END;

-- @block create `node_type` table
-- @conn bomctl
-- @label Initialize `node_type` table
-- #----------------------------------------------#
-- #----[ NodeType enum ]-------------------------#
-- #----------------------------------------------#
-- @label Create table
CREATE TABLE IF NOT EXISTS node_type (
    id   INTEGER PRIMARY KEY,
    name TEXT
);

-- @label Populate values
INSERT OR IGNORE INTO node_type (id, name) VALUES (0, 'PACKAGE');
INSERT OR IGNORE INTO node_type (id, name) VALUES (1, 'FILE');

-- @label Define triggers
CREATE TRIGGER IF NOT EXISTS tr_node_type_before_update
BEFORE UPDATE ON node_type
BEGIN
    SELECT RAISE(IGNORE);
END;

CREATE TRIGGER IF NOT EXISTS tr_node_type_before_insert
BEFORE INSERT ON node_type
BEGIN
    SELECT RAISE(IGNORE);
END;

-- @block create `edge_type` table
-- @conn bomctl
-- @label Initialize `edge_type` table
-- #----------------------------------------------#
-- #----[ EdgeType enum ]-------------------------#
-- #----------------------------------------------#
-- @label Create table
CREATE TABLE IF NOT EXISTS edge_type (
    id   INTEGER PRIMARY KEY,
    name TEXT
);

-- @label Populate values
INSERT OR IGNORE INTO edge_type (id, name) VALUES (0, 'UNKNOWN');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (1, 'amends');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (2, 'ancestor');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (3, 'buildDependency');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (4, 'buildTool');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (5, 'contains');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (6, 'contained_by');;
INSERT OR IGNORE INTO edge_type (id, name) VALUES (7, 'copy');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (8, 'dataFile');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (9, 'dependencyManifest');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (10, 'dependsOn');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (11, 'dependencyOf');;
INSERT OR IGNORE INTO edge_type (id, name) VALUES (12, 'descendant');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (13, 'describes');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (14, 'describedBy');;
INSERT OR IGNORE INTO edge_type (id, name) VALUES (15, 'devDependency');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (16, 'devTool');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (17, 'distributionArtifact');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (18, 'documentation');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (19, 'dynamicLink');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (20, 'example');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (21, 'expandedFromArchive');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (22, 'fileAdded');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (23, 'fileDeleted');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (24, 'fileModified');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (25, 'generates');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (26, 'generatedFrom');;
INSERT OR IGNORE INTO edge_type (id, name) VALUES (27, 'metafile');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (28, 'optionalComponent');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (29, 'optionalDependency');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (30, 'other');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (31, 'packages');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (32, 'patch');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (33, 'prerequisite');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (34, 'prerequisiteFor');;
INSERT OR IGNORE INTO edge_type (id, name) VALUES (35, 'providedDependency');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (36, 'requirementFor');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (37, 'runtimeDependency');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (38, 'specificationFor');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (39, 'staticLink');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (40, 'test');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (41, 'testCase');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (42, 'testDependency');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (43, 'testTool');
INSERT OR IGNORE INTO edge_type (id, name) VALUES (44, 'variant');

-- @label Define triggers
CREATE TRIGGER IF NOT EXISTS tr_edge_type_before_update
BEFORE UPDATE ON edge_type
BEGIN
    SELECT RAISE(IGNORE);
END;

CREATE TRIGGER IF NOT EXISTS tr_edge_type_before_insert
BEFORE INSERT ON edge_type
BEGIN
    SELECT RAISE(IGNORE);
END;
