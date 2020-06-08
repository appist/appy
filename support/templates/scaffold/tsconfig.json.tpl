{
  "compilerOptions": {
    "allowSyntheticDefaultImports": true,
    "baseUrl": "web/src",
    "emitDecoratorMetadata": true,
    "esModuleInterop": true,
    "experimentalDecorators": true,
    "forceConsistentCasingInFileNames": true,
    "importHelpers": true,
    "isolatedModules": true,
    "lib": ["dom", "esnext", "es6"],
    "module": "esnext",
    "moduleResolution": "node",
    "noImplicitAny": true,
    "noImplicitReturns": true,
    "noImplicitThis": true,
    "rootDir": "web/src",
    "sourceMap": true,
    "strict": true,
    "strictBindCallApply": true,
    "strictFunctionTypes": true,
    "strictPropertyInitialization": true,
    "strictNullChecks": true,
    "target": "esnext",
    "types": ["jest", "node", "@pyoner/svelte-types"],
    "typeRoots": ["node_modules/@types", "web/src/typings"],
    "paths": {
      "@assets": ["../../assets"],
      "@": ["*"]
    }
  },
  "exclude": ["node_modules"],
  "include": ["web/src/**/*"]
}
