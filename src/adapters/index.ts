import { AbcDebitAdapter } from './abc-debit';
import { BocomCreditAdapter } from './bocom-credit';
import { BocomDebitAdapter } from './bocom-debit';
import { CgbCreditAdapter } from './cgb-credit'
import { CmbCreditAdapter } from './cmb-credit'
import { CmbDebitAdapter } from './cmb-debit'
import { BocDebitAdapter } from './boc-debit'
import { BocCreditAdapter } from './boc-credit'
import { Adapter } from './types';
import { compareKey } from '../utils';

export const AdapterList = [
    AbcDebitAdapter,
    CgbCreditAdapter,
    CmbCreditAdapter,
    CmbDebitAdapter,
    BocomCreditAdapter,
    BocomDebitAdapter,
    BocDebitAdapter,
    BocCreditAdapter
].sort(
    (a, b) => compareKey(a.key, b.key)
);

export const AdapterKeys = AdapterList.map((a) => a.key);

export const AdapterMap = AdapterList.reduce<Record<Adapter['key'], Adapter>>(
    (pre, cur) => {
        pre[cur.key] = cur;
        return pre;
    },
    {}
);
