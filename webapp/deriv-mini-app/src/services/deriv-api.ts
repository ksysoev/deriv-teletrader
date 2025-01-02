import DerivAPIBasic from '@deriv/deriv-api/dist/DerivAPIBasic';

class DerivAPIService {
    private static instance: DerivAPIService;
    private api: DerivAPIBasic | null = null;
    private connection: WebSocket | null = null;
    private tickSubscription: any = null;
    private isConnecting: boolean = false;
    private connectionPromise: Promise<void> | null = null;
    private readonly APP_ID = '1089'; // Deriv app_id

    private constructor() {
        this.connect();
    }

    static getInstance(): DerivAPIService {
        if (!DerivAPIService.instance) {
            DerivAPIService.instance = new DerivAPIService();
        }
        return DerivAPIService.instance;
    }

    private connect(): Promise<void> {
        if (this.connectionPromise) {
            return this.connectionPromise;
        }

        if (this.connection?.readyState === WebSocket.OPEN && this.api) {
            return Promise.resolve();
        }

        this.connectionPromise = new Promise((resolve, reject) => {
            try {
                console.log('Connecting to WebSocket...');
                this.connection = new WebSocket(`wss://ws.binaryws.com/websockets/v3?app_id=${this.APP_ID}`);
                
                this.connection.onopen = () => {
                    console.log('WebSocket connected, initializing API...');
                    try {
                        this.api = new DerivAPIBasic({
                            connection: this.connection!,
                            app_id: this.APP_ID,
                        });
                        console.log('API initialized successfully');
                        this.isConnecting = false;
                        this.connectionPromise = null;
                        resolve();
                    } catch (error) {
                        console.error('Failed to initialize API:', error);
                        reject(error);
                    }
                };

                this.connection.onclose = () => {
                    console.log('WebSocket disconnected');
                    this.api = null;
                    this.connection = null;
                    this.isConnecting = false;
                    this.connectionPromise = null;
                    this.tickSubscription = null;
                    // Reconnect after 1 second
                    setTimeout(() => this.connect(), 1000);
                };

                this.connection.onerror = (error) => {
                    console.error('WebSocket error:', error);
                    this.isConnecting = false;
                    this.connectionPromise = null;
                    reject(error);
                };

                this.connection.onmessage = (msg) => {
                    try {
                        const data = JSON.parse(msg.data);
                        if (data.error) {
                            console.error('API error:', data.error);
                        }
                    } catch (error) {
                        console.error('Error parsing message:', error);
                    }
                };
            } catch (error) {
                console.error('Error creating WebSocket:', error);
                this.isConnecting = false;
                this.connectionPromise = null;
                reject(error);
            }
        });

        return this.connectionPromise;
    }

    async subscribeTicks(symbol: string, callback: (tick: any) => void) {
        try {
            console.log('Ensuring connection is established...');
            await this.connect();

            if (!this.api) {
                throw new Error('API not initialized');
            }

            if (this.tickSubscription) {
                console.log('Unsubscribing from previous subscription...');
                await this.tickSubscription.unsubscribe();
                this.tickSubscription = null;
            }

            console.log('Subscribing to ticks for', symbol);
            const request = {
                ticks: symbol,
                subscribe: 1
            };

            this.tickSubscription = await this.api.subscribe(request);

            if (!this.tickSubscription) {
                throw new Error('Failed to create subscription');
            }

            console.log('Setting up tick callback...');
            this.tickSubscription.onUpdate((response: any) => {
                try {
                    if (response.error) {
                        console.error('Tick error:', response.error);
                        return;
                    }
                    callback(response);
                } catch (error) {
                    console.error('Error in tick callback:', error);
                }
            });

            console.log('Successfully subscribed to ticks for', symbol);
        } catch (error) {
            console.error('Error subscribing to ticks:', error);
            // Try to reconnect
            this.connection?.close();
            throw error;
        }
    }

    async unsubscribeTicks() {
        if (this.tickSubscription) {
            try {
                console.log('Unsubscribing from ticks...');
                await this.tickSubscription.unsubscribe();
                this.tickSubscription = null;
                console.log('Successfully unsubscribed from ticks');
            } catch (error) {
                console.error('Error unsubscribing from ticks:', error);
            }
        }
    }

    disconnect() {
        console.log('Disconnecting...');
        this.unsubscribeTicks();
        if (this.connection?.readyState === WebSocket.OPEN) {
            this.connection.close();
        }
        this.api = null;
        this.connection = null;
        console.log('Disconnected');
    }
}

export default DerivAPIService;
