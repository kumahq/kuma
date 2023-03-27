import{d as g,ah as m,o as s,c as p,f as o,Q as i,b as a,u as c,x as u,_ as v,i as y,k as h,w as l,E as _,a as k,y as x,a9 as $,L as O,N}from"./index-be86debd.js";const w={class:"onboarding-heading"},B={class:"onboarding-title"},I={key:0,class:"onboarding-description"},q=g({__name:"OnboardingHeading",setup(n){const e=m();return(t,d)=>(s(),p("div",w,[o("h1",B,[i(t.$slots,"title",{},void 0,!0)]),a(),c(e).description?(s(),p("div",I,[i(t.$slots,"description",{},void 0,!0)])):u("",!0)]))}});const j=v(q,[["__scopeId","data-v-fdf98afa"]]),C={class:"onboarding-actions"},V={class:"button-list"},E=g({__name:"OnboardingNavigation",props:{shouldAllowNext:{type:Boolean,required:!1,default:!0},showSkip:{type:Boolean,required:!1,default:!0},nextStep:{type:String,required:!0},previousStep:{type:String,required:!1,default:""},nextStepTitle:{type:String,required:!1,default:"Next"},lastStep:{type:Boolean,required:!1,default:!1}},setup(n){const e=n,t=y();function d(){t.dispatch("onboarding/completeOnboarding")}function b(f){t.dispatch("onboarding/changeStep",f)}return(f,r)=>(s(),p("div",C,[e.previousStep?(s(),h(c(_),{key:0,appearance:"secondary",to:{name:e.previousStep},"data-testid":"onboarding-previous-button",onClick:r[0]||(r[0]=S=>b(e.previousStep))},{default:l(()=>[a(`
      Back
    `)]),_:1},8,["to"])):u("",!0),a(),o("div",V,[e.showSkip?(s(),h(c(_),{key:0,appearance:"outline","data-testid":"onboarding-skip-button",to:{name:"home"},onClick:d},{default:l(()=>[a(`
        Skip setup
      `)]),_:1})):u("",!0),a(),k(c(_),{disabled:!e.shouldAllowNext,appearance:e.lastStep?"creation":"primary",to:{name:e.lastStep?"home":e.nextStep},"data-testid":"onboarding-next-button",onClick:r[1]||(r[1]=S=>e.lastStep?d():b(e.nextStep))},{default:l(()=>[a(x(e.nextStepTitle),1)]),_:1},8,["disabled","appearance","to"])])]))}});const F=v(E,[["__scopeId","data-v-da07ae4c"]]),H=n=>(O("data-v-18dc3352"),n=n(),N(),n),P={class:"onboarding-container"},T={class:"onboarding-container__header"},A={class:"onboarding-container__inner-content"},z={class:"mt-4"},D=H(()=>o("div",{class:"background-image"},null,-1)),L=g({__name:"OnboardingPage",props:{withImage:{type:Boolean,required:!1,default:!1}},setup(n){const e=n;return(t,d)=>(s(),p("div",null,[o("div",P,[o("div",T,[i(t.$slots,"header",{},void 0,!0)]),a(),o("div",{class:$(["onboarding-container__content",{"onboarding-container__content--with-image":e.withImage}])},[o("div",A,[i(t.$slots,"content",{},void 0,!0)])],2),a(),o("div",z,[i(t.$slots,"navigation",{},void 0,!0)])]),a(),D]))}});const G=v(L,[["__scopeId","data-v-18dc3352"]]);export{j as O,F as a,G as b};
