import{d as c,O as b,o as s,c as d,l as n,a as i,e as o,m,q as u,_,r as v,p as g,w as p,b as f,t as S,n as y}from"./index-DHF7c3Xb.js";const k={class:"onboarding-heading"},$={class:"onboarding-title","data-testid":"onboarding-header"},x={key:0,class:"onboarding-description"},h=c({__name:"OnboardingHeading",setup(r){const t=b();return(a,e)=>(s(),d("div",k,[n("h1",$,[i(a.$slots,"title",{},void 0,!0)]),e[0]||(e[0]=o()),m(t).description?(s(),d("div",x,[i(a.$slots,"description",{},void 0,!0)])):u("",!0)]))}}),T=_(h,[["__scopeId","data-v-505a1a6e"]]),w={class:"onboarding-actions"},B={class:"button-list"},N=c({__name:"OnboardingNavigation",props:{shouldAllowNext:{type:Boolean,required:!1,default:!0},showSkip:{type:Boolean,required:!1,default:!0},nextStep:{type:String,required:!0},previousStep:{type:String,required:!1,default:""},nextStepTitle:{type:String,required:!1,default:"Next"},lastStep:{type:Boolean,required:!1,default:!1}},setup(r){const t=r;return(a,e)=>{const l=v("XAction");return s(),d("div",w,[t.previousStep?(s(),g(l,{key:0,appearance:"secondary",to:{name:t.previousStep},"data-testid":"onboarding-previous-button"},{default:p(()=>e[0]||(e[0]=[o(`
      Back
    `)])),_:1},8,["to"])):u("",!0),e[3]||(e[3]=o()),n("div",B,[t.showSkip?(s(),g(l,{key:0,appearance:"tertiary","data-testid":"onboarding-skip-button",to:{name:"home"}},{default:p(()=>e[1]||(e[1]=[o(`
        Skip setup
      `)])),_:1})):u("",!0),e[2]||(e[2]=o()),f(l,{disabled:!t.shouldAllowNext,appearance:"primary",to:{name:t.lastStep?"home":t.nextStep},"data-testid":"onboarding-next-button"},{default:p(()=>[o(S(t.nextStepTitle),1)]),_:1},8,["disabled","to"])])])}}}),H=_(N,[["__scopeId","data-v-91497fd4"]]),q={class:"onboarding-container"},O={class:"onboarding-container__header"},C={class:"onboarding-container__inner-content"},I={class:"mt-4"},A=c({__name:"OnboardingPage",props:{withImage:{type:Boolean,required:!1,default:!1}},setup(r){const t=r;return(a,e)=>(s(),d("div",null,[n("div",q,[n("div",O,[i(a.$slots,"header",{},void 0,!0)]),e[0]||(e[0]=o()),n("div",{class:y(["onboarding-container__content",{"onboarding-container__content--with-image":t.withImage}])},[n("div",C,[i(a.$slots,"content",{},void 0,!0)])],2),e[1]||(e[1]=o()),n("div",I,[i(a.$slots,"navigation",{},void 0,!0)])]),e[2]||(e[2]=o()),e[3]||(e[3]=n("div",{class:"background-image"},null,-1))]))}}),P=_(A,[["__scopeId","data-v-cd1eae59"]]);export{P as O,T as a,H as b};