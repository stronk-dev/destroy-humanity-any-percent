import Decimal from "break_eternity.js";

export type DecimalSource = Decimal | string | number;

export function sumGeometricSeries(
  count: number,
  base: DecimalSource,
  ratio: DecimalSource,
  owned: number,
): Decimal {
  const start = new Decimal(base).mul(new Decimal(ratio).pow(owned));
  return start
    .mul(Decimal.sub(1, new Decimal(ratio).pow(count)))
    .div(Decimal.sub(1, ratio));
}

export function affordGeometricSeries(
  cash: DecimalSource,
  base: DecimalSource,
  ratio: DecimalSource,
  owned: number,
): Decimal {
  const actualStart = new Decimal(base).mul(new Decimal(ratio).pow(owned));
  return new Decimal(cash)
    .div(actualStart)
    .mul(new Decimal(ratio).sub(1))
    .add(1)
    .log10()
    .div(new Decimal(ratio).log10())
    .floor();
}
