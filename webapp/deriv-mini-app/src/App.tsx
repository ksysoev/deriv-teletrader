import React, { useState } from 'react';
import styled from 'styled-components';
import { LineChart, Line, ResponsiveContainer, YAxis, CartesianGrid } from 'recharts';

const AppContainer = styled.div`
  background-color: #0e0e0e;
  min-height: 100vh;
  color: white;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
  position: relative;
  overflow-y: auto;
  padding-bottom: 160px; // Increased space for fixed buttons
`;

const ScrollContainer = styled.div`
  height: 100%;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
`;

const Header = styled.header`
  padding: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #1e1e1e;
  background-color: #0e0e0e;
  position: sticky;
  top: 0;
  z-index: 10;
`;

const Balance = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 500;
`;

const BalanceIcon = styled.div`
  width: 32px;
  height: 32px;
  background-color: #2a2a2a;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #00a99c;
`;

const ChartSection = styled.div`
  padding: 16px;
  height: 320px; // Increased height
  background-color: #141414;
`;

const ChartHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
`;

const IconButton = styled.button`
  background: none;
  border: none;
  color: #999;
  font-size: 20px;
  cursor: pointer;
  padding: 4px 8px;
  margin-left: 8px;
  transition: color 0.2s;

  &:hover {
    color: white;
  }
`;

const TradingSection = styled.div`
  padding: 16px;
`;

const InputGroup = styled.div`
  margin-bottom: 16px;
`;

const InputLabel = styled.div`
  color: #999;
  font-size: 14px;
  margin-bottom: 8px;
`;

const Input = styled.input`
  width: 100%;
  padding: 12px;
  background-color: #1e1e1e;
  border: none;
  border-radius: 4px;
  color: white;
  font-size: 16px;
  box-sizing: border-box;

  &:focus {
    background-color: #2a2a2a;
  }
`;

const TradeControls = styled.div`
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 16px;
`;

const TradeControl = styled.div`
  padding: 12px;
  background-color: #1e1e1e;
  border-radius: 4px;
  text-align: center;
`;

const TradeButtons = styled.div`
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 16px;
  background-color: #0e0e0e;
  border-top: 1px solid #1e1e1e;
  z-index: 10;
`;

const TradeButton = styled.button<{ $variant: 'up' | 'down' }>`
  padding: 16px;
  border: none;
  border-radius: 4px;
  color: white;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  background-color: ${props => props.$variant === 'up' ? '#00a99c' : '#ff444f'};
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  transition: transform 0.2s, opacity 0.2s;

  &:hover {
    transform: translateY(-2px);
    opacity: 0.9;
  }

  &:active {
    transform: translateY(0);
    opacity: 1;
  }
`;

const ButtonIcon = styled.div<{ $variant: 'up' | 'down' }>`
  font-size: 24px;
  transform: ${props => props.$variant === 'up' ? 'rotate(-45deg)' : 'rotate(45deg)'};
`;

const Value = styled.div<{ $trend?: 'up' | 'down' }>`
  text-align: center;
  font-size: 24px;
  margin-top: 32px;
  padding-bottom: 8px;
  color: ${props => props.$trend === 'up' ? '#00a99c' : props.$trend === 'down' ? '#ff444f' : 'white'};
`;

// Mock data for the chart
const mockChartData = Array.from({ length: 20 }, (_, i) => ({
  value: 1000 + Math.sin(i * 0.5) * 200 + Math.random() * 50
}));

function App() {
  const [amount, setAmount] = useState('10.00');
  const [multiplier, setMultiplier] = useState('x10');

  return (
    <AppContainer>
      <ScrollContainer>
        <Header>
          <Balance>
            <BalanceIcon>D</BalanceIcon>
            10,251.38 USD
          </Balance>
          <div>Multipliers ▼</div>
        </Header>

        <ChartSection>
          <ChartHeader>
            <div>Volatility 100 Index</div>
            <div>
              <IconButton>📊</IconButton>
              <IconButton>⭐</IconButton>
              <IconButton>ℹ️</IconButton>
            </div>
          </ChartHeader>
          <ResponsiveContainer width="100%" height="70%">
            <LineChart data={mockChartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#1e1e1e" opacity={0.3} />
              <YAxis hide={true} domain={['auto', 'auto']} />
              <Line
                type="monotone"
                dataKey="value"
                stroke="#00a99c"
                dot={false}
                strokeWidth={2}
              />
            </LineChart>
          </ResponsiveContainer>
          <Value $trend="up">
            1452.94 ▲
          </Value>
        </ChartSection>

        <TradingSection>
          <h2>Set order</h2>
          <InputGroup>
            <InputLabel>Amount (USD)</InputLabel>
            <Input
              type="text"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
            />
          </InputGroup>
          <InputGroup>
            <InputLabel>Multiplier</InputLabel>
            <Input
              type="text"
              value={multiplier}
              onChange={(e) => setMultiplier(e.target.value)}
            />
          </InputGroup>

          <TradeControls>
            <TradeControl>
              <InputLabel>TP (USD)</InputLabel>
              <div>-</div>
            </TradeControl>
            <TradeControl>
              <InputLabel>SL (USD)</InputLabel>
              <div>-</div>
            </TradeControl>
            <TradeControl>
              <InputLabel>DC</InputLabel>
              <div>-</div>
            </TradeControl>
          </TradeControls>
        </TradingSection>
      </ScrollContainer>

      <TradeButtons>
        <TradeButton $variant="up">
          <ButtonIcon $variant="up">↗</ButtonIcon>
          Up
          <div>{amount} USD</div>
        </TradeButton>
        <TradeButton $variant="down">
          <ButtonIcon $variant="down">↘</ButtonIcon>
          Down
          <div>{amount} USD</div>
        </TradeButton>
      </TradeButtons>
    </AppContainer>
  );
}

export default App;
