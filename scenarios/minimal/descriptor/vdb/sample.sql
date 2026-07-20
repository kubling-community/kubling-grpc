CREATE FOREIGN TABLE TYPE_COVERAGE (

    ID string(64) NOT NULL,

    STRING_VALUE string(100),
    VARBINARY_VALUE varbinary,
    CHAR_VALUE char(1),

    BOOLEAN_VALUE boolean,

    BYTE_VALUE byte,
    SHORT_VALUE short,
    INTEGER_VALUE integer,
    LONG_VALUE long,

    BIGINTEGER_VALUE biginteger,

    FLOAT_VALUE float,
    DOUBLE_VALUE double,
    DECIMAL_VALUE bigdecimal,

    DATE_VALUE date,
    TIME_VALUE time,
    TIMESTAMP_VALUE timestamp,

    JSON_VALUE json,

    PRIMARY KEY(ID)

)
OPTIONS (
    NAMEINSOURCE '"PUBLIC"."TYPE_COVERAGE"',
    UPDATABLE TRUE,
    "kbl_rel:fqn" 'catalog=PORTABLE-1/schema=PUBLIC/table=TYPE_COVERAGE',
    "kbl_rel:source_type" 'BASE TABLE'
);