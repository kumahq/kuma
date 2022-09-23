import{_ as c,o as l,f,A as b,p as x,k as w,cQ as h,b as O,w as a,d as s,h as e,r}from"./index.b4c50a9a.js";import{O as k}from"./OnboardingNavigation.2099801b.js";import{O as $,a as y}from"./OnboardingPage.17b77734.js";const B={name:"ServiceBox",props:{active:{type:Boolean,default:!1}},emits:["clicked"],computed:{classes(){return["box",{"box--active":this.active}]}}};function S(n,t,p,m,u,o){return l(),f("div",{"data-testid":"box",class:x(o.classes),onClick:t[0]||(t[0]=i=>n.$emit("clicked"))},[b(n.$slots,"default",{},void 0,!0)],2)}const A=c(B,[["render",S],["__scopeId","data-v-612c6d89"]]),C=""+new URL("new-service-demo.bff0792e.svg",import.meta.url).href,N=""+new URL("new-service-manually.5bec5301.svg",import.meta.url).href,M={name:"AddNewServices",components:{OnboardingNavigation:k,OnboardingHeading:$,OnboardingPage:y,ServiceBox:A},metaInfo(){return{title:"Add new services"}},computed:{...w({onboardingMode:"onboarding/getMode"}),nextStep(){return this.mode==="manually"?"onboarding-completed":"onboarding-add-services-code"},mode:{get(){return this.onboardingMode},set(n){this.update(n)}}},methods:{...h({update:"onboarding/UPDATE_MODE"})}},P={class:"h-full w-full flex justify-evenly items-center"},D=e("div",null,[e("img",{src:C}),e("div",{class:"ml-3"},[e("p",{class:"uppercase font-bold tracking-wider"}," Demo app "),e("p",null,"Counter application")])],-1),E=e("div",{class:"cursor-pointer"},[e("img",{src:N}),e("div",{class:"ml-3"},[e("p",{class:"uppercase font-bold tracking-wider"}," Manually "),e("p",null,"After this wizard")])],-1);function H(n,t,p,m,u,o){const i=r("OnboardingHeading"),d=r("ServiceBox"),v=r("OnboardingNavigation"),_=r("OnboardingPage");return l(),O(_,null,{header:a(()=>[s(i,{title:"Add services"})]),content:a(()=>[e("div",P,[s(d,{active:o.mode==="demo",class:"cursor-pointer",onClicked:t[0]||(t[0]=g=>n.update("demo"))},{default:a(()=>[D]),_:1},8,["active"]),s(d,{active:o.mode==="manually",class:"cursor-pointer",onClicked:t[1]||(t[1]=g=>n.update("manually"))},{default:a(()=>[E]),_:1},8,["active"])])]),navigation:a(()=>[s(v,{"next-step":o.nextStep,"previous-step":"onboarding-create-mesh"},null,8,["next-step"])]),_:1})}const L=c(M,[["render",H]]);export{L as default};
