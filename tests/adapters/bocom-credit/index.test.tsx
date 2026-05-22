import { vi, beforeEach, test, describe, expect } from 'vitest';

import pageHtmlText from './page.html?raw';
import resultCsvText from './result.csv?raw';

const wrappedHeaderHtmlText = `
  <table id="table3">
    <tbody>
      <tr>
        <td>交易日<br>期</td>
        <td>记账日<br>期</td>
        <td>交易说明</td>
        <td>交易币种/金额</td>
        <td>入账币种/金额</td>
      </tr>
      <tr>
        <td colspan="5">人民币账户明细</td>
      </tr>
      <tr>
        <td>2025-<br>01-13</td>
        <td>2025-<br>01-13</td>
        <td>信用卡还款<br>跨行自助转账还款-张三<br>0000 跨行自助还款</td>
        <td>CNY 47.00</td>
        <td>CNY 47.00</td>
      </tr>
      <tr>
        <td colspan="5">以下是您的消费、取现及其他费用明细</td>
      </tr>
      <tr>
        <td>2024-12-21</td>
        <td>2024-12-21</td>
        <td>分期扣款 1/6期-商户分期-示例商户C-分6期</td>
        <td>CNY 599.83</td>
        <td>CNY 599.83</td>
      </tr>
    </tbody>
  </table>
`;

const mockPostalMimeHtml = (html: string) => {
  vi.doMock('postal-mime', () => ({
    default: {
      parse: vi.fn(() => ({
        html,
      }))
    }
  }));
};

describe('adapter for bocom credit', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.resetModules();
  });

  test('converter', async () => {
    mockPostalMimeHtml(pageHtmlText);
    const { BocomCreditAdapter } = await import('../../../src/adapters/bocom-credit');

    const mockFile = new File([], 'bocom_credit.eml');
    const csvText = await BocomCreditAdapter.converter(mockFile);

    expect(csvText).toEqual(resultCsvText);
  });

  test('支持表头换行且表头在 tbody 中的明细表', async () => {
    mockPostalMimeHtml(wrappedHeaderHtmlText);
    const { BocomCreditAdapter } = await import('../../../src/adapters/bocom-credit');

    const mockFile = new File([], 'bocom_credit_wrapped_header.eml');
    const csvText = await BocomCreditAdapter.converter(mockFile);

    expect(csvText).toEqual([
      '交易日期,记账日期,交易说明,交易币种/金额,入账币种/金额',
      '2025-01-13,2025-01-13,信用卡还款 跨行自助转账还款-张三 0000 跨行自助还款,CNY 47.00,CNY 47.00',
      '2024-12-21,2024-12-21,分期扣款 1/6期-商户分期-示例商户C-分6期,CNY 599.83,CNY 599.83',
    ].join('\n'));
  });
});
