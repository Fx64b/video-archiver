import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Button } from '../button'

describe('Button', () => {
    it('should render button with text', () => {
        render(<Button>Click me</Button>)
        expect(screen.getByText('Click me')).toBeInTheDocument()
    })

    it('should handle click events', async () => {
        const user = userEvent.setup()
        const handleClick = jest.fn()

        render(<Button onClick={handleClick}>Click me</Button>)

        await user.click(screen.getByText('Click me'))

        expect(handleClick).toHaveBeenCalledTimes(1)
    })

    it('should apply default variant and size', () => {
        render(<Button>Default Button</Button>)
        const button = screen.getByText('Default Button')

        expect(button).toHaveClass('bg-primary')
        expect(button).toHaveClass('h-9')
    })

    it('should apply destructive variant', () => {
        render(<Button variant="destructive">Delete</Button>)
        const button = screen.getByText('Delete')

        expect(button).toHaveClass('bg-destructive')
    })

    it('should apply outline variant', () => {
        render(<Button variant="outline">Outline</Button>)
        const button = screen.getByText('Outline')

        expect(button).toHaveClass('border')
    })

    it('should apply secondary variant', () => {
        render(<Button variant="secondary">Secondary</Button>)
        const button = screen.getByText('Secondary')

        expect(button).toHaveClass('bg-secondary')
    })

    it('should apply ghost variant', () => {
        render(<Button variant="ghost">Ghost</Button>)
        const button = screen.getByText('Ghost')

        expect(button).toHaveClass('hover:bg-accent')
    })

    it('should apply link variant', () => {
        render(<Button variant="link">Link</Button>)
        const button = screen.getByText('Link')

        expect(button).toHaveClass('text-primary')
        expect(button).toHaveClass('underline-offset-4')
    })

    it('should apply small size', () => {
        render(<Button size="sm">Small</Button>)
        const button = screen.getByText('Small')

        expect(button).toHaveClass('h-8')
    })

    it('should apply large size', () => {
        render(<Button size="lg">Large</Button>)
        const button = screen.getByText('Large')

        expect(button).toHaveClass('h-10')
    })

    it('should apply icon size', () => {
        render(<Button size="icon" aria-label="Icon button">X</Button>)
        const button = screen.getByLabelText('Icon button')

        expect(button).toHaveClass('size-9')
    })

    it('should apply custom className', () => {
        render(<Button className="custom-class">Custom</Button>)
        const button = screen.getByText('Custom')

        expect(button).toHaveClass('custom-class')
    })

    it('should be disabled when disabled prop is true', async () => {
        const user = userEvent.setup()
        const handleClick = jest.fn()

        render(<Button disabled onClick={handleClick}>Disabled</Button>)
        const button = screen.getByText('Disabled')

        expect(button).toBeDisabled()
        expect(button).toHaveClass('disabled:opacity-50')

        await user.click(button)
        expect(handleClick).not.toHaveBeenCalled()
    })

    it('should render as child component when asChild is true', () => {
        render(
            <Button asChild>
                <a href="/test">Link Button</a>
            </Button>
        )

        const link = screen.getByText('Link Button')
        expect(link.tagName).toBe('A')
        expect(link).toHaveAttribute('href', '/test')
    })

    it('should have data-slot attribute', () => {
        render(<Button>Test</Button>)
        const button = screen.getByText('Test')

        expect(button).toHaveAttribute('data-slot', 'button')
    })
})
