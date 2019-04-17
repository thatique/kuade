--
-- Cassandra keyspace with table creation
--
--
CREATE KEYSPACE IF NOT EXISTS ${keyspace} WITH replication = ${replication};

CREATE TYPE IF NOT EXISTS ${keyspace}.geo_point {
  latitude double,
  longitude doube,
}

CREATE TYPE IF NOT EXISTS ${keyspace}.address {
  address  text,
  address2 text,
  city     text,
  state    text,
  zipcode  text,
  point    geo_point,
}

-- table for users
CREATE TABLE IF NOT EXISTS ${keyspace}.users (
  id         bigint,
  slug       text,
  name       text,
  email      text,
  icon       text,
  role       int,
  status     int,
  bio        text,
  age        int,
  address    address,
  category   text,
  budget     int,
  created_at timestamp,
  PRIMARY KEY (id)
);

-- table for users to enable us query a user by slug
CREATE TABLE IF NOT EXISTS ${keyspace}.users_by_slug (
  slug       text,
  id         bigint,
  name       text,
  email      text,
  icon       text,
  role       int,
  status     int,
  bio        text,
  age        int,
  address    address,
  category   text,
  budget     int,
  created_at timestamp,
  PRIMARY KEY (slug)
);

CREATE TABLE IF NOT EXISTS ${keyspace}.user_credentials (
  email        text,
  user_id      bigint,
  password     blob,
  enabled      boolean,
  created_at   timestamp,
  last_signin  timestamp,

  PRIMARY KEY (email)
);

-- table to store user's session using sersan lib
CREATE TABLE IF NOT EXISTS ${keyspace}.sessions (
    id          text,
    auth_id     text,
    values      blob,
    created_at  timestamp,
    accessed_at timestamp,

    PRIMARY KEY (id)
)
    WITH compaction = {
        'compaction_window_size': '1',
        'compaction_window_unit': 'HOURS',
        'class': 'org.apache.cassandra.db.compaction.TimeWindowCompactionStrategy'
    }
    AND dclocal_read_repair_chance = 0.0
    AND default_time_to_live = ${session_ttl}
    AND speculative_retry = 'NONE'
    and gc_grace_seconds = 10800;

CREATE TABLE IF NOT EXISTS ${keyspace}.session_auth_index (
    auth_id text,
    id      text,

    PRIMARY KEY (auth_id, id)
)
    WITH compaction = {
        'compaction_window_size': '1',
        'compaction_window_unit': 'HOURS',
        'class': 'org.apache.cassandra.db.compaction.TimeWindowCompactionStrategy'
    }
    AND dclocal_read_repair_chance = 0.0
    AND default_time_to_live = ${session_ttl}
    AND speculative_retry = 'NONE'
    and gc_grace_seconds = 10800;
