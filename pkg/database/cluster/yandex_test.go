package cluster

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_expandYql(t *testing.T) {
	for _, tt := range []struct {
		in  string
		out string
	}{
		{
			in: yqlInsertAccount,
			//nolint:lll
			out: "DECLARE $bic AS String; DECLARE $ban AS String; DECLARE $balance AS Int64;\nINSERT INTO `stroppy/account` (bic, ban, balance) VALUES ($bic, $ban, $balance);\n",
		},
		{
			in: yqlTransfer,
			//nolint:lll
			out: "DECLARE $transfer_id AS String;\nDECLARE $src_bic AS String;\nDECLARE $src_ban AS String;\nDECLARE $dst_bic AS String;\nDECLARE $dst_ban AS String;\nDECLARE $amount AS Int64;\nDECLARE $state AS String;\n\n$shared_select = (\n    SELECT\n        bic,\n        ban,\n        Ensure(balance - $amount, balance >= $amount, 'INSUFFICIENT_FUNDS') AS balance\n    FROM `stroppy/account`\n    WHERE bic = $src_bic AND ban = $src_ban\n    UNION ALL\n    SELECT\n        bic,\n        ban,\n        balance + $amount AS balance\n    FROM `stroppy/account`\n    WHERE bic = $dst_bic AND ban = $dst_ban\n);\n\nDISCARD SELECT Ensure(2, cnt=2, 'MISSING_ACCOUNTS')\nFROM (SELECT COUNT(*) AS cnt FROM $shared_select);\n\nUPSERT INTO `stroppy/account`\nSELECT * FROM $shared_select;\n\nUPSERT INTO `stroppy/transfer` (transfer_id, src_bic, src_ban, dst_bic, dst_ban, amount, state)\nVALUES ($transfer_id, $src_bic, $src_ban, $dst_bic, $dst_ban, $amount, $state);\n",
		},
		{
			in: yqlUpsertTransfer,
			//nolint:lll
			out: "DECLARE $transfer_id AS String;\nDECLARE $src_bic AS String;\nDECLARE $src_ban AS String;\nDECLARE $dst_bic AS String;\nDECLARE $dst_ban AS String;\nDECLARE $amount AS Int64;\nDECLARE $state AS String;\nUPSERT INTO `stroppy/transfer` (\n    transfer_id,\n    src_bic,\n    src_ban,\n    dst_bic,\n    dst_ban,\n    amount,\n    state\n)\nVALUES (\n    $transfer_id,\n    $src_bic,\n    $src_ban,\n    $dst_bic,\n    $dst_ban,\n    $amount,\n    $state\n);\n",
		},
		{
			in: yqlSelectBalanceAccount,
			//nolint:lll
			out: "DECLARE $bic AS String; DECLARE $ban AS String;\n\nSELECT balance, CAST(0 AS Int64) AS pending\nFROM `stroppy/account`\nWHERE bic = $bic AND ban = $ban\n",
		},
		{
			in: yqlSelectSrcDstAccount,
			//nolint:lll
			out: "DECLARE $src_bic AS String;\nDECLARE $src_ban AS String;\nDECLARE $dst_bic AS String;\nDECLARE $dst_ban AS String;\n\nSELECT 1 AS srcdst, balance\nFROM `stroppy/account`\nWHERE bic = $src_bic AND ban = $src_ban\nUNION ALL\nSELECT 2 AS srcdst, balance\nFROM `stroppy/account`\nWHERE bic = $dst_bic AND ban = $dst_ban;\n",
		},
		{
			in: yqlUpsertSrcDstAccount,
			//nolint:lll
			out: "DECLARE $src_bic AS String;\nDECLARE $src_ban AS String;\nDECLARE $dst_bic AS String;\nDECLARE $dst_ban AS String;\nDECLARE $amount AS Int64;\n\n$shared_select = (\n    SELECT\n        bic,\n        ban,\n        balance - $amount AS balance\n    FROM `stroppy/account`\n    WHERE bic = $src_bic AND ban = $src_ban\n    UNION ALL\n    SELECT\n        bic,\n        ban,\n        balance + $amount AS balance\n    FROM `stroppy/account`\n    WHERE bic = $dst_bic AND ban = $dst_ban\n);\n\nUPSERT INTO `stroppy/account`\nSELECT * FROM $shared_select;\n",
		},
	} {
		t.Run("", func(t *testing.T) {
			require.Equal(t, tt.out, expandYql(tt.in))
		})
	}
}
