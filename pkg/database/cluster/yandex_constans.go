package cluster

const (
	insertYdbTransfer = `
DECLARE $params AS Struct<
    transfer_id:String, 
    src_bic:String,
    src_ban:String,
    dst_bic:String,
    dst_ban:String,
    amount:Int64,
    state:String
>;
UPSERT INTO %s (
    transfer_id,
    src_bic,
    src_ban,
    dst_bic,
    dst_ban,
    amount,
    state
)
VALUES (
    $params.transfer_id,
    $params.src_bic,
    $params.src_ban,
    $params.dst_bic,
    $params.dst_ban,
    $params.amount,
    $params.state
);`
	srcAndDstYdbSelect = `
DECLARE $params AS Struct<
    src_bic:String,
    src_ban:String,
    dst_bic:String,
    dst_ban:String,
>;
SELECT 
    bic,
    ban,
    balance
FROM %s
WHERE bic = $params.src_bic AND ban = $params.src_ban
UNION ALL
SELECT 
    bic,
    ban,
    balance
FROM %s
WHERE bic = $params.dst_bic AND ban = $params.dst_ban;
`
	unifiedTransfer = `
DECLARE $params AS Struct<
    src_bic:String,
    src_ban:String,
    dst_bic:String,
    dst_ban:String,
    amount:Int64,
>;
$shared_select = (
    SELECT 
        bic,
        ban,
        balance - $params.amount AS balance
    FROM %s
    WHERE bic = $params.src_bic AND ban = $params.src_ban
    UNION ALL
    SELECT 
        bic,
        ban,
        balance + $params.amount AS balance
    FROM %s
    WHERE bic = $params.dst_bic AND ban = $params.dst_ban
);

UPDATE %s ON
SELECT * FROM $shared_select;
`

	singleStatementTransfer = `
DECLARE $transfer_id AS String;
DECLARE $src_bic AS String;
DECLARE $src_ban AS String;
DECLARE $dst_bic AS String;
DECLARE $dst_ban AS String;
DECLARE $amount AS Int64;
DECLARE $state AS String;

$count = SELECT COUNT(*) AS rows FROM (
    SELECT 1 FROM ` + "`%s/account`" + `
    WHERE (bic = $src_bic AND ban = $src_ban)
       OR (bic = $dst_bic AND ban = $dst_ban)
);

$acc_rows = (
    SELECT bic, ban, balance - $amount AS balance
    FROM ` + "`%s/account`" + `
    WHERE bic = $src_bic AND ban = $src_ban AND $count = 2
    UNION ALL
    SELECT bic, ban, balance + $amount as balance
    FROM ` + "`%s/account`" + `
    WHERE bic = $dst_bic AND ban = $dst_ban AND $count = 2
);

$trans_rows = (
    SELECT $transfer_id AS transfer_id,
      $src_bic AS src_bic,
      $src_ban AS src_ban,
      $dst_bic AS dst_bic,
      $dst_ban AS dst_ban,
      $amount AS amount,
      $state AS state
    FROM (SELECT 1 AS x) y
    WHERE $count = 2
);

UPSERT INTO ` + "`%s/transfer`" + `
  SELECT * FROM $trans_rows;

UPSERT INTO ` + "`%s/account`" + `
  SELECT * FROM $acc_rows;

SELECT balance FROM $acc_rows;
`
)
