-- ------------------------------------------------------------------------
-- SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
-- SPDX-FileName: internal/pkg/db/scripts/tables.sql
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

-- @block Create `metadata` table
-- @conn bomctl
-- @label Initialize `metadata` table
CREATE TABLE IF NOT EXISTS metadata (
    id      TEXT CHECK (LENGTH(id) == 45) PRIMARY KEY,
    version TEXT,
    name    TEXT,
    date    TEXT CHECK (date IS date(date)),
    comment TEXT
);

-- @block Create `tool` table
-- @conn bomctl
-- @label Initialize `tool` table
CREATE TABLE IF NOT EXISTS tool (
    name        TEXT,
    version     TEXT,
    vendor      TEXT,

    PRIMARY KEY (name, version, vendor)
);

-- @block Create `metadata_tool` table
-- @conn bomctl
-- @label Initialize `metadata_tool` table
CREATE TABLE IF NOT EXISTS metadata_tool (
    metadata_id  TEXT NOT NULL REFERENCES metadata (id),
    tool_name    TEXT REFERENCES tool (name),
    tool_version TEXT REFERENCES tool (version),
    tool_vendor  TEXT REFERENCES tool (vendor),

    PRIMARY KEY (metadata_id, tool_name, tool_version, tool_vendor)
);

-- @block Create `person` table
-- @conn bomctl
-- @label Initialize `person` table
CREATE TABLE IF NOT EXISTS person (
    name        TEXT,
    is_org      INTEGER CHECK(is_org IN (TRUE, FALSE)),
    email       TEXT,
    url         TEXT,
    phone       TEXT,

    PRIMARY KEY (name, is_org, email, url, phone)
);

-- @block Create `metadata_contact` table
-- @conn bomctl
-- @label Initialize `metadata_contact` table
CREATE TABLE IF NOT EXISTS metadata_contact (
    metadata_id   TEXT NOT NULL REFERENCES metadata (id),
    person_name   TEXT          REFERENCES person (name),
    person_is_org INTEGER       REFERENCES person (is_org),
    person_email  TEXT          REFERENCES person (email),
    person_url    TEXT          REFERENCES person (url),
    person_phone  TEXT          REFERENCES person (phone),

    PRIMARY KEY (metadata_id, person_name, person_is_org, person_email, person_url, person_phone)
);

-- @block Create `person_contact` table
-- @conn bomctl
-- @label Initialize `person_contact` table
CREATE TABLE IF NOT EXISTS person_contact (
    person_name    TEXT    REFERENCES person (name),
    person_is_org  INTEGER REFERENCES person (is_org),
    person_email   TEXT    REFERENCES person (email),
    person_url     TEXT    REFERENCES person (url),
    person_phone   TEXT    REFERENCES person (phone),
    contact_name   TEXT    REFERENCES metadata_contact (person_name),
    contact_is_org INTEGER REFERENCES metadata_contact (person_is_org),
    contact_email  TEXT    REFERENCES metadata_contact (person_email),
    contact_url    TEXT    REFERENCES metadata_contact (person_url),
    contact_phone  TEXT    REFERENCES metadata_contact (person_phone),

    PRIMARY KEY (
        person_name,
        person_is_org,
        person_email,
        person_url,
        person_phone,
        contact_name,
        contact_is_org,
        contact_email,
        contact_url,
        contact_phone
    )
);

-- @block Create `document_type` table
-- @conn bomctl
-- @label Initialize `document_type` table
CREATE TABLE IF NOT EXISTS document_type (
    metadata_id TEXT NOT NULL REFERENCES metadata (id),
    type        INTEGER REFERENCES sbom_type (id),
    name        TEXT,
    description TEXT,
    PRIMARY KEY (metadata_id, type, name, description)
);

-- @block Create `external_reference` table
-- @conn bomctl
-- @label Initialize `external_reference` table
CREATE TABLE IF NOT EXISTS external_reference (
    url       TEXT,
    comment   TEXT,
    authority TEXT,
    type      INTEGER REFERENCES external_reference_type (id),

    PRIMARY KEY (url, comment,authority, type)
);

-- @block Create `external_reference_hashes` table
-- @conn bomctl
-- @label Initialize `external_reference_hashes` table
CREATE TABLE IF NOT EXISTS external_reference_hashes (
    external_reference_url       TEXT,
    external_reference_comment   TEXT,
    external_reference_authority TEXT,
    external_reference_type      INTEGER REFERENCES external_reference_type (id),
    hash_algorithm               INTEGER REFERENCES hash_algorithm (id),
    hash_data                    TEXT,

    PRIMARY KEY (
        external_reference_url,
        external_reference_comment,
        external_reference_authority,
        external_reference_type,
        hash_algorithm,
        hash_data
    ),

    CONSTRAINT fk_external_references_hashes FOREIGN KEY (
        external_reference_url,
        external_reference_comment,
        external_reference_authority,
        external_reference_type
    ) REFERENCES external_references (url, comment, authority, type)
);

-- @block Create `node` table
-- @conn bomctl
-- @label Initialize `node` table
CREATE TABLE IF NOT EXISTS node (
    id                TEXT CHECK(LENGTH(id) == 45) PRIMARY KEY,
    type              INTEGER REFERENCES node_type (id),
    name              TEXT,
    version           TEXT,
    file_name         TEXT,
    url_home          TEXT,
    url_download      TEXT,
    license_concluded TEXT,
    license_comments  TEXT,
    copyright         TEXT,
    source_info       TEXT,
    comment           TEXT,
    summary           TEXT,
    description       TEXT,
    release_date      TEXT CHECK (release_date IS date(release_date)),
    build_date        TEXT CHECK (build_date IS date(build_date)),
    valid_until_date  TEXT CHECK (valid_until_date IS date(valid_until_date))
);

-- @block Create `node_license` table
-- @conn bomctl
-- @label Initialize `node_license` table
CREATE TABLE IF NOT EXISTS node_license (
    node_id TEXT REFERENCES node (id),
    license TEXT,

    PRIMARY KEY (node_id, license)
);

-- @block Create `node_attribution` table
-- @conn bomctl
-- @label Initialize `node_attribution` table
CREATE TABLE IF NOT EXISTS node_attribution (
    node_id     TEXT REFERENCES node (id),
    attribution TEXT,

    PRIMARY KEY (node_id, attribution)
);

-- @block Create `node_supplier` table
-- @conn bomctl
-- @label Initialize `node_supplier` table
CREATE TABLE IF NOT EXISTS node_supplier (
    node_id         TEXT    REFERENCES node (id),
    supplier_name   TEXT    REFERENCES person (name),
    supplier_is_org INTEGER REFERENCES person (is_org),
    supplier_email  TEXT    REFERENCES person (email),
    supplier_url    TEXT    REFERENCES person (url),
    supplier_phone  TEXT    REFERENCES person (phone),

    PRIMARY KEY (node_id, supplier_name, supplier_is_org, supplier_email, supplier_url, supplier_phone)
);

-- @block Create `node_originator` table
-- @conn bomctl
-- @label Initialize `node_originator` table
CREATE TABLE IF NOT EXISTS node_originator (
    node_id           TEXT    REFERENCES node (id),
    originator_name   TEXT    REFERENCES person (name),
    originator_is_org INTEGER REFERENCES person (is_org),
    originator_email  TEXT    REFERENCES person (email),
    originator_url    TEXT    REFERENCES person (url),
    originator_phone  TEXT    REFERENCES person (phone),

    PRIMARY KEY (node_id, originator_name, originator_is_org, originator_email, originator_url, originator_phone)
);

-- @block Create `node_external_reference` table
-- @conn bomctl
-- @label Initialize `node_external_reference` table
CREATE TABLE IF NOT EXISTS node_external_reference (
    node_id                      TEXT    REFERENCES node (id),
    external_reference_url       TEXT    REFERENCES external_reference (url),
    external_reference_comment   TEXT    REFERENCES external_reference (comment),
    external_reference_authority TEXT    REFERENCES external_reference (authority),
    external_reference_type      INTEGER REFERENCES external_reference (type),

    PRIMARY KEY (
      node_id,
      external_reference_url,
      external_reference_comment,
      external_reference_authority,
      external_reference_type
    )
);

-- @block Create `node_file_type` table
-- @conn bomctl
-- @label Initialize `node_file_type` table
CREATE TABLE IF NOT EXISTS node_file_type (
    node_id   TEXT REFERENCES node (id),
    file_type TEXT,

    PRIMARY KEY (node_id, file_type)
);

-- @block Create `node_software_identifier` table
-- @conn bomctl
-- @label Initialize `node_software_identifier` table
CREATE TABLE IF NOT EXISTS node_software_identifier (
    node_id                   TEXT    REFERENCES node (id),
    software_identifier_type  INTEGER REFERENCES software_identifier (id),
    software_identifier_value TEXT,

    PRIMARY KEY (node_id, software_identifier_type, software_identifier_value)
);

-- @block Create `node_hash` table
-- @conn bomctl
-- @label Initialize `node_hash` table
CREATE TABLE IF NOT EXISTS node_hash (
    node_id             TEXT    REFERENCES node (id),
    hash_algorithm_type INTEGER REFERENCES hash_algorithm (id),
    hash_data           TEXT,

    PRIMARY KEY (node_id, hash_algorithm_type, hash_data)
);

-- @block Create `node_primary_purpose` table
-- @conn bomctl
-- @label Initialize `node_primary_purpose` table
CREATE TABLE IF NOT EXISTS node_primary_purpose (
    node_id         TEXT    REFERENCES node (id),
    primary_purpose INTEGER REFERENCES purpose (id),

    PRIMARY KEY (node_id, primary_purpose)
);

-- @block Create `edge` table
-- @conn bomctl
-- @label Initialize `edge` table
CREATE TABLE IF NOT EXISTS edge (
    id        INTEGER PRIMARY KEY,
    type      INTEGER NOT NULL REFERENCES edge_type (id),
    edge_from TEXT
);

-- @block Create `edge_from_to` table
-- @conn bomctl
-- @label Initialize `edge_from_to` table
CREATE TABLE IF NOT EXISTS edge_from_to (
    edge_id INTEGER NOT NULL REFERENCES edge (id),
    edge_to TEXT    NOT NULL,

    PRIMARY KEY (edge_id, edge_to)
);

-- @block Create `node_list` table
-- @conn bomctl
-- @label Initialize `node_list` table
CREATE TABLE IF NOT EXISTS node_list (
    id INTEGER PRIMARY KEY
);

-- @block Create `node_list_node` table
-- @conn bomctl
-- @label Initialize `node_list_node` table
CREATE TABLE IF NOT EXISTS node_list_node (
    node_list_id INTEGER REFERENCES node_list (id),
    node_id      TEXT    REFERENCES node (id),

    PRIMARY KEY (node_list_id, node_id)
);

-- @block Create `node_list_edge` table
-- @conn bomctl
-- @label Initialize `node_list_edge` table
CREATE TABLE IF NOT EXISTS node_list_edge (
    node_list_id INTEGER REFERENCES node_list (id),
    edge_id      INTEGER REFERENCES edge (id),

    PRIMARY KEY (node_list_id, edge_id)
);

-- @block Create `node_list_root_element` table
-- @conn bomctl
-- @label Initialize `node_list_root_element` table
CREATE TABLE IF NOT EXISTS node_list_root_element (
    node_list_id INTEGER REFERENCES node_list (id),
    root_element TEXT    NOT NULL,

    PRIMARY KEY (node_list_id, root_element)
);

-- @block Define table indexes
-- @conn bomctl
-- @label Create indexes
CREATE INDEX IF NOT EXISTS idx_metadata_tool ON metadata_tool (metadata_id);
CREATE INDEX IF NOT EXISTS idx_metadata_contact ON metadata_contact (metadata_id);
CREATE INDEX IF NOT EXISTS idx_document_type ON document_type (metadata_id);
CREATE INDEX IF NOT EXISTS idx_node_license ON node_license (node_id);
CREATE INDEX IF NOT EXISTS idx_node_attribution ON node_attribution (node_id);
CREATE INDEX IF NOT EXISTS idx_node_supplier ON node_supplier (node_id);
CREATE INDEX IF NOT EXISTS idx_node_originator ON node_originator (node_id);
CREATE INDEX IF NOT EXISTS idx_node_external_reference ON node_external_reference (node_id);
CREATE INDEX IF NOT EXISTS idx_node_file_type ON node_file_type (node_id);
CREATE INDEX IF NOT EXISTS idx_node_software_identifier ON node_software_identifier (node_id);
CREATE INDEX IF NOT EXISTS idx_node_hash ON node_hash (node_id);
CREATE INDEX IF NOT EXISTS idx_node_primary_purpose ON node_primary_purpose (node_id);
