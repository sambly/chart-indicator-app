// time ISO 8601 > UTC
export function convTime(time:string):any {
    const date = new Date(time);
    return Math.floor(date.getTime() / 1000);  // Преобразование в Unix timestamp в секунды 
}