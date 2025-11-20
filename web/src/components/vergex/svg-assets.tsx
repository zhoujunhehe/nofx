// SVG Assets for VergeX Landing Page
// These are extracted from Figma design

export const SecurityIcon = () => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="96"
    height="96"
    viewBox="0 0 96 96"
    fill="none"
  >
    <path
      d="M58.9961 45.7887V39.1887C58.9961 36.2713 57.8371 33.4733 55.7742 31.4106C53.7115 29.3477 50.9135 28.1887 47.9961 28.1887C45.0787 28.1887 42.2807 29.3477 40.218 31.4106C38.1551 33.4733 36.9961 36.2713 36.9961 39.1887V45.7887"
      stroke="white"
      strokeWidth="3"
    />
    <path
      d="M32.5996 61.1922C32.5996 62.9425 33.295 64.6213 34.5328 65.859C35.7705 67.0968 37.4493 67.7922 39.1996 67.7922H56.7996C58.5499 67.7922 60.2288 67.0968 61.4665 65.859C62.7042 64.6213 63.3996 62.9425 63.3996 61.1922V47.9922H32.5996V61.1922Z"
      fill="white"
      stroke="white"
      strokeWidth="3"
    />
    <path
      d="M28.1946 80.9899H23.7946C21.4607 80.9899 19.2224 80.0628 17.5721 78.4123C15.9218 76.7621 14.9946 74.5238 14.9946 72.1899V23.7899C14.9946 21.456 15.9218 19.2177 17.5721 17.5673C19.2224 15.917 21.4607 14.9899 23.7946 14.9899H72.1946C74.5286 14.9899 76.7669 15.917 78.4171 17.5673C80.0676 19.2177 80.9946 21.456 80.9946 23.7899V72.1899C80.9946 74.5238 80.0676 76.7621 78.4171 78.4123C76.7669 80.0628 74.5286 80.9899 72.1946 80.9899H67.7946"
      fill="url(#paint0_linear_1_975)"
      fillOpacity="0.24"
    />
    <path
      d="M28.1946 80.9899H23.7946C21.4607 80.9899 19.2224 80.0628 17.5721 78.4123C15.9218 76.7621 14.9946 74.5239 14.9946 72.1899V23.7899C14.9946 21.456 15.9218 19.2177 17.5721 17.5673C19.2224 15.917 21.4607 14.9899 23.7946 14.9899H72.1946C74.5286 14.9899 76.7669 15.917 78.4171 17.5673C80.0676 19.2177 80.9946 21.456 80.9946 23.7899V72.1899C80.9946 74.5239 80.0676 76.7621 78.4171 78.4123C76.7669 80.0628 74.5286 80.9899 72.1946 80.9899H67.7946"
      stroke="white"
      strokeWidth="3"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      opacity="0.4"
      d="M36.9961 80.9843H58.9961"
      stroke="white"
      strokeWidth="3"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <defs>
      <linearGradient
        id="paint0_linear_1_975"
        x1="47.9946"
        y1="14.9899"
        x2="47.9946"
        y2="80.9899"
        gradientUnits="userSpaceOnUse"
      >
        <stop stopColor="white" stopOpacity="0" />
        <stop offset="1" stopColor="white" />
      </linearGradient>
    </defs>
  </svg>
)

export const NaturalLanguageIcon = () => (
  <svg
    width="64"
    height="64"
    viewBox="0 0 64 64"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    {/* Background rounded square */}
    <rect width="64" height="64" rx="16" fill="white" fillOpacity="0.04" />
    <rect
      x="0.5"
      y="0.5"
      width="63"
      height="63"
      rx="15.5"
      stroke="white"
      strokeOpacity="0.1"
    />

    {/* Speech bubble - main circle */}
    <circle cx="32" cy="28" r="12" stroke="white" strokeWidth="2" />

    {/* Speech bubble - tail pointing down-left */}
    <path d="M24 36L20 42L26 40" fill="white" />

    {/* Pause icon inside bubble */}
    <rect x="28" y="24" width="2" height="8" rx="1" fill="white" />
    <rect x="34" y="24" width="2" height="8" rx="1" fill="white" />
  </svg>
)

export const IntelligentBacktestIcon = () => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="64"
    height="64"
    viewBox="0 0 64 64"
    fill="none"
  >
    <path
      d="M31.9953 52.7904C43.4828 52.7904 52.7953 43.4779 52.7953 31.9904C52.7953 20.5029 43.4828 11.1904 31.9953 11.1904C20.5078 11.1904 11.1953 20.5029 11.1953 31.9904C11.1953 43.4779 20.5078 52.7904 31.9953 52.7904Z"
      fill="url(#paint0_linear_1_1172)"
      fillOpacity="0.24"
      stroke="white"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      opacity="0.4"
      d="M32.1996 29C33.0482 29 33.862 29.3371 34.4621 29.9372C35.0622 30.5373 35.3993 31.3511 35.3993 32.1996C35.3993 33.0482 35.0622 33.862 34.4621 34.4621C33.862 35.0622 33.0482 35.3993 32.1996 35.3993C31.3511 35.3993 30.5373 35.0622 29.9372 34.4621C29.3371 33.862 29 33.0482 29 32.1996C29 31.3511 29.3371 30.5373 29.9372 29.9372C30.5373 29.3371 31.3511 29 32.1996 29Z"
      fill="#080808"
      stroke="white"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      d="M32 7.99329C32.8486 7.99329 33.6626 8.33042 34.2627 8.93055C34.8629 9.53066 35.2 10.3446 35.2 11.1933C35.2 12.042 34.8629 12.8559 34.2627 13.456C33.6626 14.0562 32.8486 14.3933 32 14.3933C31.1514 14.3933 30.3374 14.0562 29.7373 13.456C29.1371 12.8559 28.8 12.042 28.8 11.1933C28.8 10.3446 29.1371 9.53066 29.7373 8.93055C30.3374 8.33042 31.1514 7.99329 32 7.99329ZM32 49.5933C32.8486 49.5933 33.6626 49.9304 34.2627 50.5306C34.8629 51.1307 35.2 51.9446 35.2 52.7933C35.2 53.6419 34.8629 54.4559 34.2627 55.056C33.6626 55.6562 32.8486 55.9933 32 55.9933C31.1514 55.9933 30.3374 55.6562 29.7373 55.056C29.1371 54.4559 28.8 53.6419 28.8 52.7933C28.8 51.9446 29.1371 51.1307 29.7373 50.5306C30.3374 49.9304 31.1514 49.5933 32 49.5933ZM52.8 28.7933C53.6486 28.7933 54.4626 29.1304 55.0627 29.7306C55.6629 30.3307 56 31.1446 56 31.9933C56 32.8419 55.6629 33.6558 55.0627 34.256C54.4626 34.8562 53.6486 35.1933 52.8 35.1933C51.9514 35.1933 51.1374 34.8562 50.5373 34.256C49.9371 33.6558 49.6 32.8419 49.6 31.9933C49.6 31.1446 49.9371 30.3307 50.5373 29.7306C51.1374 29.1304 51.9514 28.7933 52.8 28.7933ZM11.2 28.7933C12.0487 28.7933 12.8626 29.1304 13.4627 29.7306C14.0629 30.3307 14.4 31.1446 14.4 31.9933C14.4 32.8419 14.0629 33.6558 13.4627 34.256C12.8626 34.8562 12.0487 35.1933 11.2 35.1933C10.3513 35.1933 9.53738 34.8562 8.93726 34.256C8.33714 33.6558 8 32.8419 8 31.9933C8 31.1446 8.33714 30.3307 8.93726 29.7306C9.53738 29.1304 10.3513 28.7933 11.2 28.7933Z"
      fill="#080808"
      stroke="white"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <defs>
      <linearGradient
        id="paint0_linear_1_1172"
        x1="31.9953"
        y1="11.1904"
        x2="31.9953"
        y2="52.7904"
        gradientUnits="userSpaceOnUse"
      >
        <stop stopColor="white" stopOpacity="0" />
        <stop offset="1" stopColor="white" />
      </linearGradient>
    </defs>
  </svg>
)

export const AlwaysOnIcon = () => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="64"
    height="64"
    viewBox="0 0 64 64"
    fill="none"
  >
    <path
      d="M46.7169 17.5928H17.2769C16.0887 17.5928 14.9492 18.0424 14.109 18.8424C13.2689 19.6425 12.7969 20.7278 12.7969 21.8595V46.9261C12.7969 48.0577 13.2689 49.143 14.109 49.9432C14.9492 50.7432 16.0887 51.1928 17.2769 51.1928H46.7169C47.905 51.1928 49.0446 50.7432 49.8847 49.9432C50.7249 49.143 51.1969 48.0577 51.1969 46.9261V21.8595C51.1969 20.7278 50.7249 19.6425 49.8847 18.8424C49.0446 18.0424 47.905 17.5928 46.7169 17.5928Z"
      fill="url(#paint0_linear_1_1180)"
      fillOpacity="0.24"
      stroke="white"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      d="M12.7969 22.3928C12.7969 21.1198 13.2689 19.8989 14.109 18.9987C14.9492 18.0985 16.0887 17.5928 17.2769 17.5928H46.7169C47.905 17.5928 49.0446 18.0985 49.8847 18.9987C50.7249 19.8989 51.1969 21.1198 51.1969 22.3928V23.9928H12.7969V22.3928Z"
      fill="white"
      fillOpacity="0.4"
      stroke="white"
      strokeWidth="2"
    />
    <path
      d="M19.1953 12.7965V17.5965M44.7953 12.7965V17.5965"
      stroke="white"
      strokeWidth="2"
      strokeLinecap="round"
    />
    <path
      d="M20.7969 33.5965V30.3965H23.9969V33.5965H20.7969ZM20.7969 43.1965V39.9965H23.9969V43.1965H20.7969ZM30.3969 33.5965V30.3965H33.5969V33.5965H30.3969ZM30.3969 43.1965V39.9965H33.5969V43.1965H30.3969Z"
      fill="white"
      stroke="white"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      opacity="0.4"
      d="M40 33.5965V30.3965H43.2V33.5965H40Z"
      fill="white"
      stroke="white"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <defs>
      <linearGradient
        id="paint0_linear_1_1180"
        x1="35.1969"
        y1="17.5928"
        x2="35.1969"
        y2="55.9928"
        gradientUnits="userSpaceOnUse"
      >
        <stop stopColor="white" stopOpacity="0" />
        <stop offset="1" stopColor="white" />
      </linearGradient>
    </defs>
  </svg>
)

export const VergeXTriangle = () => (
  <svg width="40" height="40" viewBox="0 0 40 40" fill="none" xmlns="http://www.w3.org/2000/svg">
    <rect width="40" height="40" rx="8" fill="url(#paint0_linear_vx)" transform="rotate(45 20 20)"/>
    <path d="M20 10L10 30H30L20 10Z" fill="white" transform="rotate(45 20 20)"/>
    <defs>
      <linearGradient id="paint0_linear_vx" x1="0" y1="0" x2="40" y2="40" gradientUnits="userSpaceOnUse">
        <stop stopColor="#998cff"/>
        <stop offset="1" stopColor="#7d6ee5"/>
      </linearGradient>
    </defs>
  </svg>
)

export const GithubIcon = () => (
  <svg width="100" height="100" viewBox="0 0 100 100" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path fillRule="evenodd" clipRule="evenodd" d="M50 10C27.9086 10 10 28.3198 10 50.9091C10 69.0309 21.8629 84.4309 38.4357 90C40.6429 90.4091 41.4286 89.0436 41.4286 87.8836V80.6836C28.9214 83.4109 26.2857 75.0164 26.2857 75.0164C24.2857 69.9854 21.2857 68.6364 21.2857 68.6364C17.1429 65.7073 21.5714 65.7673 21.5714 65.7673C26.1429 66.0818 28.5 70.5691 28.5 70.5691C32.5714 77.3745 39.2143 75.3509 41.6429 74.2073C42.0714 71.2382 43.2143 69.2182 44.5 68.1145C34.2143 66.9873 23.4286 63.2218 23.4286 46.5564C23.4286 42.0218 25 38.3009 28.5714 35.3982C28.0714 34.2745 26.5 29.5691 29.0714 23.3145C29.0714 23.3145 32.9286 22.1109 41.4286 27.8909C45.0714 26.9018 49 26.4073 52.9286 26.3873C56.8571 26.4073 60.7857 26.9018 64.4286 27.8909C72.9286 22.1109 76.7857 23.3145 76.7857 23.3145C79.3571 29.5691 77.7857 34.2745 77.2857 35.3982C80.8571 38.3009 82.4286 42.0218 82.4286 46.5564C82.4286 63.2618 71.6429 66.9673 61.3571 68.0909C63 69.5364 64.4286 72.3436 64.4286 76.6418V87.8836C64.4286 89.0636 65.2143 90.4291 67.4286 90C84 84.4109 95.8571 69.0109 95.8571 50.9091C95.8571 28.3198 78.0714 10 56 10H50Z" fill="white"/>
  </svg>
)

export const ProductsArchitectureDiagram = () => (
  <svg width="589" height="620" viewBox="0 0 589 620" fill="none" xmlns="http://www.w3.org/2000/svg">
    {/* User Layer - Top */}
    <g transform="translate(0, 0)">
      <path d="M100 170 L489 170 L544 254 L489 338 L100 338 L45 254 Z" fill="rgba(153, 140, 255, 0.1)" stroke="rgba(153, 140, 255, 0.3)" strokeWidth="1.5"/>
      <text x="294" y="264" fill="white" fontSize="20" textAnchor="middle" fontFamily="Red Hat Text" letterSpacing="0.8">USER</text>
    </g>

    {/* VergeX Layer - Middle Left */}
    <g transform="translate(0, 140)">
      <path d="M100 170 L489 170 L544 254 L489 338 L100 338 L45 254 Z" fill="rgba(153, 140, 255, 0.15)" stroke="rgba(153, 140, 255, 0.4)" strokeWidth="1.5"/>
      <text x="294" y="264" fill="white" fontSize="20" textAnchor="middle" fontFamily="Red Hat Text" letterSpacing="0.8">VERGEX</text>
    </g>

    {/* Exchange Layer - Middle Right */}
    <g transform="translate(0, 280)">
      <path d="M100 170 L489 170 L544 254 L489 338 L100 338 L45 254 Z" fill="rgba(153, 140, 255, 0.1)" stroke="rgba(153, 140, 255, 0.3)" strokeWidth="1.5"/>
      <text x="294" y="264" fill="white" fontSize="20" textAnchor="middle" fontFamily="Red Hat Text" letterSpacing="0.8">EXCHANGE</text>
    </g>

    {/* AI Models Layer - Bottom */}
    <g transform="translate(0, 420)">
      <path d="M100 30 L489 30 L544 114 L489 198 L100 198 L45 114 Z" fill="rgba(153, 140, 255, 0.15)" stroke="rgba(153, 140, 255, 0.4)" strokeWidth="1.5"/>
      <text x="294" y="124" fill="white" fontSize="20" textAnchor="middle" fontFamily="Red Hat Text" letterSpacing="0.8">AI MODELS</text>
    </g>

    {/* Connecting Lines */}
    <line x1="140" y1="238" x2="140" y2="524" stroke="rgba(255, 255, 255, 0.2)" strokeWidth="1"/>
    <line x1="307" y1="167" x2="307" y2="474" stroke="rgba(255, 255, 255, 0.2)" strokeWidth="1"/>
    <line x1="1088" y1="286" x2="1088" y2="394" stroke="rgba(255, 255, 255, 0.2)" strokeWidth="1"/>
    <line x1="1236" y1="384" x2="1236" y2="567" stroke="rgba(255, 255, 255, 0.2)" strokeWidth="1"/>
    <line x1="1363" y1="137" x2="1363" y2="476" stroke="rgba(255, 255, 255, 0.2)" strokeWidth="1"/>
  </svg>
)
