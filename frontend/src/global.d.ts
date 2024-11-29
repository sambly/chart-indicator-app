export {};

// объявил глобально иначе vite не собирает эти функции 
declare global {
  interface Window {
    buildSMA: (quote: Quote,sma:Indicator[]) => void;
    buildExtremum: (quote: Quote,highIndicator:Indicator[],lowIndicator:Indicator[]) => void;
    buidPlug: (data:any) => void;
  }
}