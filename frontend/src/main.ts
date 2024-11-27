
interface Quote {
	symbol: string;
	precision: number;
	date: string[]; // Преобразуем time.Time в строку в формате ISO
	open: number[];
	high: number[];
	low: number[];
	close: number[];
	volume: number[];
  }

export function buildChart(quote: Quote): void {

	console.log(quote);

}






