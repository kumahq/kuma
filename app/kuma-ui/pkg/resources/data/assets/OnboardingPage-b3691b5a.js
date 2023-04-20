import{d as g,a5 as S,o as s,j as p,i as n,K as i,h as o,b as c,f as u,e as m,w as _,g as y,t as k,Z as x,p as $,m as O}from"./index-5b34b65e.js";import{_ as v}from"./_plugin-vue_export-helper-c27b6911.js";import{_ as l}from"./kongponents.es-9f2d376a.js";import{u as w}from"./store-76151f83.js";const N={class:"onboarding-heading"},B={class:"onboarding-title"},I={key:0,class:"onboarding-description"},q=g({__name:"OnboardingHeading",setup(a){const e=S();return(t,d)=>(s(),p("div",N,[n("h1",B,[i(t.$slots,"title",{},void 0,!0)]),o(),c(e).description?(s(),p("div",I,[i(t.$slots,"description",{},void 0,!0)])):u("",!0)]))}});const J=v(q,[["__scopeId","data-v-fdf98afa"]]),C={class:"onboarding-actions"},V={class:"button-list"},H=g({__name:"OnboardingNavigation",props:{shouldAllowNext:{type:Boolean,required:!1,default:!0},showSkip:{type:Boolean,required:!1,default:!0},nextStep:{type:String,required:!0},previousStep:{type:String,required:!1,default:""},nextStepTitle:{type:String,required:!1,default:"Next"},lastStep:{type:Boolean,required:!1,default:!1}},setup(a){const e=a,t=w();function d(){t.dispatch("onboarding/completeOnboarding")}function f(b){t.dispatch("onboarding/changeStep",b)}return(b,r)=>(s(),p("div",C,[e.previousStep?(s(),m(c(l),{key:0,appearance:"secondary",to:{name:e.previousStep},"data-testid":"onboarding-previous-button",onClick:r[0]||(r[0]=h=>f(e.previousStep))},{default:_(()=>[o(`
      Back
    `)]),_:1},8,["to"])):u("",!0),o(),n("div",V,[e.showSkip?(s(),m(c(l),{key:0,appearance:"outline","data-testid":"onboarding-skip-button",to:{name:"home"},onClick:d},{default:_(()=>[o(`
        Skip setup
      `)]),_:1})):u("",!0),o(),y(c(l),{disabled:!e.shouldAllowNext,appearance:e.lastStep?"creation":"primary",to:{name:e.lastStep?"home":e.nextStep},"data-testid":"onboarding-next-button",onClick:r[1]||(r[1]=h=>e.lastStep?d():f(e.nextStep))},{default:_(()=>[o(k(e.nextStepTitle),1)]),_:1},8,["disabled","appearance","to"])])]))}});const L=v(H,[["__scopeId","data-v-da07ae4c"]]),P=a=>($("data-v-18dc3352"),a=a(),O(),a),T={class:"onboarding-container"},A={class:"onboarding-container__header"},j={class:"onboarding-container__inner-content"},z={class:"mt-4"},D=P(()=>n("div",{class:"background-image"},null,-1)),E=g({__name:"OnboardingPage",props:{withImage:{type:Boolean,required:!1,default:!1}},setup(a){const e=a;return(t,d)=>(s(),p("div",null,[n("div",T,[n("div",A,[i(t.$slots,"header",{},void 0,!0)]),o(),n("div",{class:x(["onboarding-container__content",{"onboarding-container__content--with-image":e.withImage}])},[n("div",j,[i(t.$slots,"content",{},void 0,!0)])],2),o(),n("div",z,[i(t.$slots,"navigation",{},void 0,!0)])]),o(),D]))}});const M=v(E,[["__scopeId","data-v-18dc3352"]]);export{M as O,J as a,L as b};
