import { render, screen } from '@testing-library/react'
import { Badge } from '../badge'

describe('Badge', () => {
    it('should render badge with text', () => {
        render(<Badge>New</Badge>)
        expect(screen.getByText('New')).toBeInTheDocument()
    })

    it('should apply default variant', () => {
        render(<Badge>Default</Badge>)
        const badge = screen.getByText('Default')

        expect(badge).toHaveClass('bg-primary')
        expect(badge).toHaveClass('border-transparent')
    })

    it('should apply secondary variant', () => {
        render(<Badge variant="secondary">Secondary</Badge>)
        const badge = screen.getByText('Secondary')

        expect(badge).toHaveClass('bg-secondary')
    })

    it('should apply destructive variant', () => {
        render(<Badge variant="destructive">Error</Badge>)
        const badge = screen.getByText('Error')

        expect(badge).toHaveClass('bg-destructive')
    })

    it('should apply outline variant', () => {
        render(<Badge variant="outline">Outline</Badge>)
        const badge = screen.getByText('Outline')

        expect(badge).toHaveClass('text-foreground')
    })

    it('should apply custom className', () => {
        render(<Badge className="custom-badge">Custom</Badge>)
        const badge = screen.getByText('Custom')

        expect(badge).toHaveClass('custom-badge')
    })

    it('should render as child component when asChild is true', () => {
        render(
            <Badge asChild>
                <a href="/test">Link Badge</a>
            </Badge>
        )

        const link = screen.getByText('Link Badge')
        expect(link.tagName).toBe('A')
        expect(link).toHaveAttribute('href', '/test')
    })

    it('should have data-slot attribute', () => {
        render(<Badge>Test</Badge>)
        const badge = screen.getByText('Test')

        expect(badge).toHaveAttribute('data-slot', 'badge')
    })

    it('should render as span by default', () => {
        render(<Badge>Test Badge</Badge>)
        const badge = screen.getByText('Test Badge')

        expect(badge.tagName).toBe('SPAN')
    })

    it('should have proper base styles', () => {
        render(<Badge>Styled</Badge>)
        const badge = screen.getByText('Styled')

        expect(badge).toHaveClass('inline-flex')
        expect(badge).toHaveClass('items-center')
        expect(badge).toHaveClass('rounded-md')
        expect(badge).toHaveClass('text-xs')
    })
})
