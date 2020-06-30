{
  "compilerOptions": {
    "allowSyntheticDefaultImports": true,
    "baseUrl": "web/src",
    "declaration": true,
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
    "types": ["jest", "node", "svelte"],
    "paths": {
      "@assets/*": ["../../assets/*"],
      "@/*": ["*"]
    }
  },
  "exclude": ["node_modules"]
}
