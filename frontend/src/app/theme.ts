import type { GlobalThemeOverrides } from 'naive-ui';

export const themeOverrides: GlobalThemeOverrides = {
  common: {
        primaryColor: '#62e6c2',
        primaryColorHover: '#2ebb96',
        primaryColorPressed: '#189b7a',
        primaryColorSuppl: '#153b34',
        bodyColor: '#0e1519',
        cardColor: '#1a272b',
        modalColor: '#1a272b',
        popoverColor: '#1a272b',
        textColorBase: '#e7f0ee',
        textColor1: '#e7f0ee',
        textColor2: '#a4b6b5',
        textColor3: '#6d8181',
        borderColor: '#26363a',
        dividerColor: '#26363a',
    borderRadius: '8px',
    borderRadiusSmall: '6px',
    fontFamily: 'IBM Plex Sans, Segoe UI, sans-serif',
    fontFamilyMono: 'JetBrains Mono, Consolas, monospace',
  },
  Button: {
    borderRadiusMedium: '6px',
    heightMedium: '34px',
    fontWeight: '600',
  },
  Input: {
    borderRadius: '6px',
    heightMedium: '36px',
  },
  Select: {
    peers: {
      InternalSelection: {
        borderRadius: '6px',
        heightMedium: '36px',
      },
    },
  },
  Modal: {
    borderRadius: '12px',
  },
  Dialog: {
    borderRadius: '10px',
  },
  Card: {
    borderRadius: '10px',
  },
  Tag: {
    borderRadius: '5px',
  },
};
