declare module '@deriv/deriv-api/dist/DerivAPIBasic' {
    interface DerivAPIConfig {
        connection: WebSocket;
        app_id?: string;
    }

    interface TickSubscription {
        unsubscribe(): Promise<void>;
        onUpdate(callback: (data: any) => void): void;
    }

    class DerivAPIBasic {
        constructor(config: DerivAPIConfig);
        subscribe(options: { ticks: string; subscribe: number }): Promise<TickSubscription>;
    }

    export default DerivAPIBasic;
}
