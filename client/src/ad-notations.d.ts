declare module "@antimatter-dimensions/notations" {
  export class StandardNotation {
    readonly name: string;
    format(value: string | number | object, places: number, placesUnder1000: number): string;
  }
}
