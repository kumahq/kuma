import{d as _,N as b,o as s,b as d,g as o,h as i,k as t,A as r,e as p,_ as u,a as h,w as c,O as l,j as f,t as v,l as m,p as S,f as y}from"./index-B1qFrf1M.js";const k={class:"onboarding-heading"},$={class:"onboarding-title","data-testid":"onboarding-header"},w={key:0,class:"onboarding-description"},x=_({__name:"OnboardingHeading",setup(a){const e=b();return(n,g)=>(s(),d("div",k,[o("h1",$,[i(n.$slots,"title",{},void 0,!0)]),t(),r(e).description?(s(),d("div",w,[i(n.$slots,"description",{},void 0,!0)])):p("",!0)]))}}),j=u(x,[["__scopeId","data-v-505a1a6e"]]),N={class:"onboarding-actions"},B={class:"button-list"},I=_({__name:"OnboardingNavigation",props:{shouldAllowNext:{type:Boolean,required:!1,default:!0},showSkip:{type:Boolean,required:!1,default:!0},nextStep:{type:String,required:!0},previousStep:{type:String,required:!1,default:""},nextStepTitle:{type:String,required:!1,default:"Next"},lastStep:{type:Boolean,required:!1,default:!1}},setup(a){const e=a;return(n,g)=>(s(),d("div",N,[e.previousStep?(s(),h(r(l),{key:0,appearance:"secondary",to:{name:e.previousStep},"data-testid":"onboarding-previous-button"},{default:c(()=>[t(`
      Back
    `)]),_:1},8,["to"])):p("",!0),t(),o("div",B,[e.showSkip?(s(),h(r(l),{key:0,appearance:"tertiary","data-testid":"onboarding-skip-button",to:{name:"home"}},{default:c(()=>[t(`
        Skip setup
      `)]),_:1})):p("",!0),t(),f(r(l),{disabled:!e.shouldAllowNext,appearance:"primary",to:{name:e.lastStep?"home":e.nextStep},"data-testid":"onboarding-next-button"},{default:c(()=>[t(v(e.nextStepTitle),1)]),_:1},8,["disabled","to"])])]))}}),z=u(I,[["__scopeId","data-v-4695c7f4"]]),O=a=>(S("data-v-41beef0f"),a=a(),y(),a),q={class:"onboarding-container"},C={class:"onboarding-container__header"},V={class:"onboarding-container__inner-content"},A={class:"mt-4"},T=O(()=>o("div",{class:"background-image"},null,-1)),H=_({__name:"OnboardingPage",props:{withImage:{type:Boolean,required:!1,default:!1}},setup(a){const e=a;return(n,g)=>(s(),d("div",null,[o("div",q,[o("div",C,[i(n.$slots,"header",{},void 0,!0)]),t(),o("div",{class:m(["onboarding-container__content",{"onboarding-container__content--with-image":e.withImage}])},[o("div",V,[i(n.$slots,"content",{},void 0,!0)])],2),t(),o("div",A,[i(n.$slots,"navigation",{},void 0,!0)])]),t(),T]))}}),D=u(H,[["__scopeId","data-v-41beef0f"]]);export{D as O,j as a,z as b};
