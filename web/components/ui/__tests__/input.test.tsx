import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Input } from '../input'

describe('Input', () => {
    it('should render input element', () => {
        render(<Input placeholder="Enter text" />)
        expect(screen.getByPlaceholderText('Enter text')).toBeInTheDocument()
    })

    it('should handle text input', async () => {
        const user = userEvent.setup()
        render(<Input placeholder="Enter text" />)

        const input = screen.getByPlaceholderText('Enter text')
        await user.type(input, 'Hello World')

        expect(input).toHaveValue('Hello World')
    })

    it('should handle onChange event', async () => {
        const user = userEvent.setup()
        const handleChange = jest.fn()

        render(<Input onChange={handleChange} placeholder="Enter text" />)

        const input = screen.getByPlaceholderText('Enter text')
        await user.type(input, 'Test')

        expect(handleChange).toHaveBeenCalled()
    })

    it('should apply custom className', () => {
        render(<Input className="custom-input" placeholder="Test" />)
        const input = screen.getByPlaceholderText('Test')

        expect(input).toHaveClass('custom-input')
    })

    it('should be disabled when disabled prop is true', async () => {
        const user = userEvent.setup()
        render(<Input disabled placeholder="Disabled input" />)

        const input = screen.getByPlaceholderText('Disabled input')

        expect(input).toBeDisabled()
        expect(input).toHaveClass('disabled:opacity-50')

        // Attempt to type should not work
        await user.type(input, 'Test')
        expect(input).toHaveValue('')
    })

    it('should handle different input types', () => {
        const { rerender } = render(<Input type="text" placeholder="Text" />)
        expect(screen.getByPlaceholderText('Text')).toHaveAttribute('type', 'text')

        rerender(<Input type="password" placeholder="Password" />)
        expect(screen.getByPlaceholderText('Password')).toHaveAttribute('type', 'password')

        rerender(<Input type="email" placeholder="Email" />)
        expect(screen.getByPlaceholderText('Email')).toHaveAttribute('type', 'email')
    })

    it('should have default styling classes', () => {
        render(<Input placeholder="Styled" />)
        const input = screen.getByPlaceholderText('Styled')

        expect(input).toHaveClass('h-10')
        expect(input).toHaveClass('w-full')
        expect(input).toHaveClass('rounded-md')
        expect(input).toHaveClass('border')
    })

    it('should forward ref correctly', () => {
        const ref = React.createRef<HTMLInputElement>()
        render(<Input ref={ref} placeholder="Ref test" />)

        expect(ref.current).toBeInstanceOf(HTMLInputElement)
        expect(ref.current?.placeholder).toBe('Ref test')
    })

    it('should handle value prop (controlled)', () => {
        const { rerender } = render(<Input value="Initial" readOnly />)
        const input = screen.getByDisplayValue('Initial')

        expect(input).toHaveValue('Initial')

        rerender(<Input value="Updated" readOnly />)
        expect(input).toHaveValue('Updated')
    })

    it('should handle defaultValue prop (uncontrolled)', () => {
        render(<Input defaultValue="Default value" />)
        const input = screen.getByDisplayValue('Default value')

        expect(input).toHaveValue('Default value')
    })
})
