/*
 * Copyright 2021 ICON Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package foundation.icon.btp.bts;

import foundation.icon.btp.lib.BSH;
import foundation.icon.btp.lib.BSHScoreClient;
import foundation.icon.btp.lib.OwnerManager;
import foundation.icon.btp.lib.OwnerManagerScoreClient;
import foundation.icon.btp.test.BTPIntegrationTest;
import foundation.icon.btp.test.MockBMCIntegrationTest;
import foundation.icon.icx.Wallet;
import foundation.icon.jsonrpc.model.TransactionResult;
import foundation.icon.score.client.DefaultScoreClient;
import foundation.icon.score.client.ScoreClient;
import foundation.icon.score.test.ScoreIntegrationTest;
import org.junit.jupiter.api.MethodOrderer;
import org.junit.jupiter.api.TestMethodOrder;

import java.util.Map;
import java.util.function.Consumer;

@TestMethodOrder(value = MethodOrderer.OrderAnnotation.class)
public interface BTSIntegrationTest extends BTPIntegrationTest {

        DefaultScoreClient btsClient = DefaultScoreClient.of(
                        System.getProperties(), Map.of(
                                        "_bmc", MockBMCIntegrationTest.mockBMCClient._address()));

        @ScoreClient
        BTS bts = new BTSScoreClient(btsClient);

        @ScoreClient
        BSH btsBSH = new BSHScoreClient(btsClient);

        @ScoreClient
        OwnerManager btsOwnerManager = new OwnerManagerScoreClient(btsClient);

        Wallet tester = ScoreIntegrationTest.getOrGenerateWallet("tester.", System.getProperties());
        DefaultScoreClient btsClientWithTester = new DefaultScoreClient(
                        btsClient.endpoint(), btsClient._nid(), tester, btsClient._address());
        BTS btsWithTester = new BTSScoreClient(btsClientWithTester);
        OwnerManager btsOwnerManagerWithTester = new OwnerManagerScoreClient(btsClientWithTester);

        static <T> Consumer<TransactionResult> eventLogChecker(
                        EventLogsSupplier<T> supplier, Consumer<T> consumer) {
                return ScoreIntegrationTest.eventLogChecker(
                                btsClient._address(), supplier, consumer);
        }

}
