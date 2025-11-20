import {
  HeroSection,
  WhyVergeXSection,
  OurProductsSection,
  SocialMediaSection,
  FooterSection,
} from './vergex-sections'

export function VergeXLandingPage() {
  return (
    <div className="bg-vergex-bg-primary min-h-screen relative overflow-hidden">
      <HeroSection />
      <WhyVergeXSection />
      <OurProductsSection />
      <SocialMediaSection />
      <FooterSection />
    </div>
  )
}

export default VergeXLandingPage
