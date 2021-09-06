package foundation.icon.btp.bmv;


import foundation.icon.btp.bmv.lib.HexConverter;
import foundation.icon.btp.bmv.lib.mpt.MPTException;
import foundation.icon.btp.bmv.lib.mpt.MerklePatriciaTree;
import org.bouncycastle.util.encoders.Hex;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.*;

import java.util.ArrayList;

import static org.junit.jupiter.api.Assertions.assertArrayEquals;
import static org.mockito.Mockito.spy;

@TestMethodOrder(MethodOrderer.OrderAnnotation.class)
public class MPTTest {
    private static MerklePatriciaTree mpt;

    @BeforeAll
    public static void setup() {
        mpt = spy(new MerklePatriciaTree());
    }

    @Test
    @Order(1)
    public void proveTransactionReceipt() throws MPTException {
        ArrayList<byte[]> proofs = new ArrayList<byte[]>(1);
        proofs.add(Hex.decode("f901cf822080b901c9f901c6a0b5e5c57f738b1874e7c9a693db757dc3106fe69009e127199842f80a447ab91382cd1bb9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000400000000000000000008000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000020010000000000000000000020000000000000000000000000000000000000020000000000040000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000040002000000000000000000000000000000000000000000000000000000000000000000000f89df89b947c5a0ce9267ed19b22f8cae653f198e3e8daf098f863a0ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3efa0000000000000000000000000019f484f4320c8fb11d0238b2b03c16fec905527a000000000000000000000000083335e0c01afac5e02ff201ba0f5979ebc4aa93fa000000000000000000000000000000000000000000000000340aad21b3b700000"));
        byte[] hash = this.getHash(proofs.get(0));
        System.out.println(Hex.toHexString(hash));
        byte[] key = Hex.decode("80");
        byte[] rootHash = Hex.decode("fe2b38c1f594b5c8cd4173c9baf34fdee48a487bc0550783bcaaa5e0403b2d98");
        byte[][] proof = new byte[1][];
        byte[] provingValue = mpt.prove(rootHash, key, proofs);
        assertArrayEquals(provingValue, HexConverter.hexStringToByteArray("f901c6a0b5e5c57f738b1874e7c9a693db757dc3106fe69009e127199842f80a447ab91382cd1bb9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000400000000000000000008000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000020010000000000000000000020000000000000000000000000000000000000020000000000040000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000040002000000000000000000000000000000000000000000000000000000000000000000000f89df89b947c5a0ce9267ed19b22f8cae653f198e3e8daf098f863a0ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3efa0000000000000000000000000019f484f4320c8fb11d0238b2b03c16fec905527a000000000000000000000000083335e0c01afac5e02ff201ba0f5979ebc4aa93fa000000000000000000000000000000000000000000000000340aad21b3b700000"));
    }

    public byte[] getHash(byte[] data) {
        return mpt.getHash(data);
    }
}


/*
Sample data from POC
for the inputs
        let blockHash = '0xc32470c2459fd607246412e23b4b4d19781c1fa24a603d47a5bc066be3b5c0af'
        let txHash    = '0xacb81623523bbabccb1638a907686bc2f3229c70e3ab51777bef0a635f3ac03f'
        let res = await verifyReceipt(txHash, blockHash)

res.receipt proof= 0xf901d2f901cf822080b901c9f901c6a0b5e5c57f738b1874e7c9a693db757dc3106fe69009e127199842f80a447ab91382cd1bb9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000400000000000000000008000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000020010000000000000000000020000000000000000000000000000000000000020000000000040000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000040002000000000000000000000000000000000000000000000000000000000000000000000f89df89b947c5a0ce9267ed19b22f8cae653f198e3e8daf098f863a0ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3efa0000000000000000000000000019f484f4320c8fb11d0238b2b03c16fec905527a000000000000000000000000083335e0c01afac5e02ff201ba0f5979ebc4aa93fa000000000000000000000000000000000000000000000000340aad21b3b700000
block hash header = c32470c2459fd607246412e23b4b4d19781c1fa24a603d47a5bc066be3b5c0af
receiptsroot = fe2b38c1f594b5c8cd4173c9baf34fdee48a487bc0550783bcaaa5e0403b2d98
receiptrootfromproof= keccak(rlp.encode(proof[0]))
txIndex/key/path = 0x0 = 80
receiptBufer = 0xf901c6a0b5e5c57f738b1874e7c9a693db757dc3106fe69009e127199842f80a447ab91382cd1bb9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000400000000000000000008000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000020010000000000000000000020000000000000000000000000000000000000020000000000040000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000040002000000000000000000000000000000000000000000000000000000000000000000000f89df89b947c5a0ce9267ed19b22f8cae653f198e3e8daf098f863a0ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3efa0000000000000000000000000019f484f4320c8fb11d0238b2b03c16fec905527a000000000000000000000000083335e0c01afac5e02ff201ba0f5979ebc4aa93fa000000000000000000000000000000000000000000000000340aad21b3b700000

// Block Header got from
 let txHash    = '0xacb81623523bbabccb1638a907686bc2f3229c70e3ab51777bef0a635f3ac03f'
let resp = await receiptProof(txHash)
rlp.ecnode(resp.header)
rlpencodedproof = f901cf822080b901c9f901c6a0b5e5c57f738b1874e7c9a693db757dc3106fe69009e127199842f80a447ab91382cd1bb9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000400000000000000000008000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000020010000000000000000000020000000000000000000000000000000000000020000000000040000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000040002000000000000000000000000000000000000000000000000000000000000000000000f89df89b947c5a0ce9267ed19b22f8cae653f198e3e8daf098f863a0ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3efa0000000000000000000000000019f484f4320c8fb11d0238b2b03c16fec905527a000000000000000000000000083335e0c01afac5e02ff201ba0f5979ebc4aa93fa000000000000000000000000000000000000000000000000340aad21b3b700000
 */