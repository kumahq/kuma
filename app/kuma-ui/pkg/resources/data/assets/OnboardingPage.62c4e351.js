import{C as c,o as a,k as o,l as t,t as d,y as p,H as s,z as g,A as u,B as h}from"./index.ea0d4a24.js";const b={name:"OnboardingHeading",props:{title:{type:String,required:!0},description:{type:String,default:""}}},f={class:"relative"},m={class:"onboarding-title"},v={key:0,class:"text-center text-lg mt-3"};function $(e,r,n,_,l,i){return a(),o("div",f,[t("h1",m,d(n.title),1),n.description?(a(),o("p",v,d(n.description),1)):p("",!0)])}const C=c(b,[["render",$],["__scopeId","data-v-40b9fa40"]]);const y={name:"OnboardingContainer",props:{withImage:{type:Boolean,default:!1}},computed:{classes(){return["onboarding-container__content",this.withImage?"onboarding-container__content--with-image":""]}}},I=e=>(u("data-v-8d8b0ce3"),e=e(),h(),e),O={class:"onboarding-container"},S={class:"onboarding-container__header"},k={class:"w-full"},w=I(()=>t("div",{class:"background-image"},null,-1));function x(e,r,n,_,l,i){return a(),o("div",null,[t("div",O,[t("div",S,[s(e.$slots,"header",{},void 0,!0)]),t("div",{class:g(i.classes)},[t("div",k,[s(e.$slots,"content",{},void 0,!0)])],2),s(e.$slots,"navigation",{},void 0,!0)]),w])}const H=c(y,[["render",x],["__scopeId","data-v-8d8b0ce3"]]);export{C as O,H as a};
