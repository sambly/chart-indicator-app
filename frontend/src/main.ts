import { createChart,CrosshairMode,CandlestickData,LineData} from 'lightweight-charts';
import {Quote,Indicator} from './types.d';
import { convTime } from './help';


window.buildSMA = buildSMA;
window.buildExtremum = buildExtremum;
window.buidPlug = buidPlug;


export function buildSMA(quote: Quote,sma:Indicator[]): void {

	const container_chart = document.getElementById('chart')!;
	const chartOptions = {
		height: 500,
		width: 700,
		autosize: true,
		layout: {
			backgroundColor: '#ffffff',
			textColor: 'rgba(33, 56, 77, 1)',
		},
		grid: {
			vertLines: {
				color: 'rgba(197, 203, 206, 0.7)',
			},
			horzLines: {
				color: 'rgba(197, 203, 206, 0.7)',
			},
		},
		crosshair: {
			mode: CrosshairMode.Normal,
		},
		timeScale: {
			timeVisible: true,
			secondsVisible: false
		},
	};

	const chart = createChart(container_chart, chartOptions);
	const candleSeries = chart.addCandlestickSeries({});
	const lineSeries = chart.addLineSeries({});  

	let candles:CandlestickData[] = [];
	for (let index = 0; index < quote.close.length; index++) {
		candles.push({ 
			time: convTime(quote.date[index]), 
			open: quote.open[index],
			high: quote.high[index],
			low: quote.low[index], 
			close: quote.close[index] 
			});
	}
	candleSeries.setData(candles);

	let lineSeriesData: LineData[] = [];
	let markersChart: any[]= [];
	for (let item of sma) {
		if (item.value!=0){
			lineSeriesData.push({time: convTime(item.date),value:item.value});
			markersChart.push({ time: convTime(item.date), position: 'inBar', color: '#0000FF', shape: 'circle' });  
		}  
	}
	lineSeries.setData(lineSeriesData);
}

export function buildExtremum(quote: Quote,highIndicator:Indicator[],lowIndicator:Indicator[]): void {

	const container_chart = document.getElementById('chart')!;
	const chartOptions = {
		height: 500,
		width: 700,
		autosize: true,
		layout: {
			backgroundColor: '#ffffff',
			textColor: 'rgba(33, 56, 77, 1)',
		},
		grid: {
			vertLines: {
				color: 'rgba(197, 203, 206, 0.7)',
			},
			horzLines: {
				color: 'rgba(197, 203, 206, 0.7)',
			},
		},
		crosshair: {
			mode: CrosshairMode.Normal,
		},
		timeScale: {
			timeVisible: true,
			secondsVisible: false
		},
	};

	const chart = createChart(container_chart, chartOptions);
	const candleSeries = chart.addCandlestickSeries({});
	const lineSeriesHigh = chart.addLineSeries({
		color: 'rgba(255, 255, 255, 0)', 
		lastValueVisible: false, 
		priceLineVisible: false
	}); 
	const lineSeriesLow = chart.addLineSeries({
		color: 'rgba(255, 255, 255, 0)', 
		lastValueVisible: false, 
		priceLineVisible: false
	}); 

	let candles:CandlestickData[] = [];
	for (let index = 0; index < quote.close.length; index++) {
		candles.push({ 
			time: convTime(quote.date[index]), 
			open: quote.open[index],
			high: quote.high[index],
			low: quote.low[index], 
			close: quote.close[index] 
			});
	}
	candleSeries.setData(candles);

	let lineSeriesDataHigh: LineData[] = [];
	let lineSeriesDataLow: LineData[] = [];
	let markersChartHigh: any[]= [];
	let markersChartLow: any[]= [];
	for (let item of highIndicator) {
		if (item.value!=0){
			lineSeriesDataHigh.push({time: convTime(item.date),value:item.value});
			markersChartHigh.push({ time: convTime(item.date), position: 'inBar', color: '#008000', shape: 'circle' });  
		}  
	}

	for (let item of lowIndicator) {
		if (item.value!=0){
			lineSeriesDataLow.push({time: convTime(item.date),value:item.value});
			markersChartLow.push({ time: convTime(item.date), position: 'inBar', color: '#FF0000', shape: 'circle' });  
		}  
	}

	lineSeriesHigh.setData(lineSeriesDataHigh);
	lineSeriesLow.setData(lineSeriesDataLow);
	lineSeriesHigh.setMarkers(markersChartHigh);
	lineSeriesLow.setMarkers(markersChartLow);

}

export function buidPlug(data:any): void {
	console.log(data);
}


