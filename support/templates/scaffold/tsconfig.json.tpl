{
  "compilerOptions": {
    "allowSyntheticDefaultImports": true,
    "baseUrl": "./web/src",
    "emitDecoratorMetadata": true,
    "esModuleInterop": true,
    "experimentalDecorators": true,
    "forceConsistentCasingInFileNames": true,
    "importHelpers": true,
    "isolatedModules": true,
    "module": "esnext",
    "moduleResolution": "node",
    "noImplicitAny": true,
    "noImplicitReturns": true,
    "noImplicitThis": true,
    "rootDir": "./web/src",
    "sourceMap": true,
    "strict": true,
    "strictBindCallApply": true,
    "strictFunctionTypes": true,
    "strictPropertyInitialization": true,
    "strictNullChecks": true,
    "target": "esnext",
    "types": ["jest", "node"],
    "typeRoots": ["node_modules/@types", "web/src/types"],
    "paths": {
      "@assets": ["../../assets"],
      "@": ["*"]
    },
    "lib": ["dom", "esnext", "es6"]
  },
  "include": ["web/src/**/*.ts", "web/src/**/*.tsx", "web/src/**/*.svelte"],
  "exclude": ["node_modules"]
}
