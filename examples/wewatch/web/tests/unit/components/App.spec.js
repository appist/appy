import '@testing-library/jest-dom/extend-expect'
import { render } from '@testing-library/svelte'
import App from '@/components/App.svelte'

test('shows proper heading when rendered', () => {
  const { getByText } = render(App, { props: { name: 'World' } })

  expect(getByText('Hello World!')).toBeInTheDocument()
})
