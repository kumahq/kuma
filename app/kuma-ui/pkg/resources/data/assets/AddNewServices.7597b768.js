import{E as l,o as p,j as b,J as x,B as O,cn as w,cO as h,i as r,c as B,w as a,a as i,l as e,b as s}from"./index.0cb244cf.js";import{O as $}from"./OnboardingNavigation.4797753a.js";import{O as k,a as y}from"./OnboardingPage.9278b395.js";const S={name:"ServiceBox",props:{active:{type:Boolean,default:!1}},emits:["clicked"],computed:{classes(){return["box",{"box--active":this.active}]}}};function N(n,t,m,u,v,o){return p(),b("div",{"data-testid":"box",class:O(o.classes),onClick:t[0]||(t[0]=d=>n.$emit("clicked"))},[x(n.$slots,"default",{},void 0,!0)],2)}const C=l(S,[["render",N],["__scopeId","data-v-a26d9032"]]),M=""+new URL("new-service-demo.bff0792e.svg",import.meta.url).href,A=""+new URL("new-service-manually.5bec5301.svg",import.meta.url).href,E={name:"AddNewServices",components:{OnboardingNavigation:$,OnboardingHeading:k,OnboardingPage:y,ServiceBox:C},computed:{...w({onboardingMode:"onboarding/getMode"}),nextStep(){return this.mode==="manually"?"onboarding-completed":"onboarding-add-services-code"},mode:{get(){return this.onboardingMode},set(n){this.update(n)}}},methods:{...h({update:"onboarding/UPDATE_MODE"})}},P={class:"h-full w-full flex justify-evenly items-center"},D=e("div",null,[e("img",{src:M}),s(),e("div",{class:"ml-3"},[e("p",{class:"uppercase font-bold tracking-wider"},`
                Demo app
              `),s(),e("p",null,"Counter application")])],-1),H=e("div",{class:"cursor-pointer"},[e("img",{src:A}),s(),e("div",{class:"ml-3"},[e("p",{class:"uppercase font-bold tracking-wider"},`
                Manually
              `),s(),e("p",null,"After this wizard")])],-1);function U(n,t,m,u,v,o){const d=r("OnboardingHeading"),c=r("ServiceBox"),_=r("OnboardingNavigation"),g=r("OnboardingPage");return p(),B(g,null,{header:a(()=>[i(d,{title:"Add services"})]),content:a(()=>[e("div",P,[i(c,{active:o.mode==="demo",class:"cursor-pointer",onClicked:t[0]||(t[0]=f=>n.update("demo"))},{default:a(()=>[D]),_:1},8,["active"]),s(),i(c,{active:o.mode==="manually",class:"cursor-pointer",onClicked:t[1]||(t[1]=f=>n.update("manually"))},{default:a(()=>[H]),_:1},8,["active"])])]),navigation:a(()=>[i(_,{"next-step":o.nextStep,"previous-step":"onboarding-create-mesh"},null,8,["next-step"])]),_:1})}const L=l(E,[["render",U]]);export{L as default};
