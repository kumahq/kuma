import{O as y,a as h,b as S}from"./OnboardingPage-5f16210e.js";import{d as m,o as v,c as B,r as C,n as k,_ as u,a as r,b as A,w as t,e as s,f as a,p as e,B as N,C as V}from"./index-646486ee.js";const R=""+new URL("new-service-demo-bff0792e.svg",import.meta.url).href,I=""+new URL("new-service-manually-5bec5301.svg",import.meta.url).href,$=m({__name:"ServiceBox",props:{active:{type:Boolean,required:!1,default:!1}},emits:["clicked"],setup(o,{emit:d}){const i=o,c=d;return(l,n)=>(v(),B("div",{class:k(["box",{"box--active":i.active}]),"data-testid":"box",onClick:n[0]||(n[0]=p=>c("clicked"))},[C(l.$slots,"default",{},void 0,!0)],2))}});const _=u($,[["__scopeId","data-v-506b07cb"]]),f=o=>(N("data-v-589f21f1"),o=o(),V(),o),O={class:"service-mode-list"},T=f(()=>e("div",{class:"service-box-content"},[e("img",{src:R}),a(),e("p",{class:"service-mode-title"},`
                  Demo app
                `),a(),e("p",null,"Counter application")],-1)),z=f(()=>e("div",{class:"service-box-content"},[e("img",{src:I}),a(),e("p",{class:"service-mode-title"},`
                  Manually
                `),a(),e("p",null,"After this wizard")],-1)),L=m({__name:"AddNewServices",props:{mode:{}},emits:["change"],setup(o,{emit:d}){const i=o,c=d;return(l,n)=>{const p=r("RouteTitle"),b=r("AppView"),g=r("RouteView");return v(),A(g,{name:"onboarding-add-services"},{default:t(({t:x})=>[s(p,{title:x("onboarding.routes.add-services.title"),render:!1},null,8,["title"]),a(),s(b,null,{default:t(()=>[s(y,null,{header:t(()=>[s(h,null,{title:t(()=>[a(`
              Add services
            `)]),_:1})]),content:t(()=>[e("div",O,[s(_,{"data-testid":"onboarding-demo",active:i.mode==="demo",onClicked:n[0]||(n[0]=w=>c("change","demo"))},{default:t(()=>[T]),_:1},8,["active"]),a(),s(_,{"data-testid":"onboarding-manually",active:i.mode==="manually",onClicked:n[1]||(n[1]=w=>c("change","manually"))},{default:t(()=>[z]),_:1},8,["active"])])]),navigation:t(()=>[s(S,{"next-step":i.mode==="manually"?"onboarding-completed":"onboarding-add-services-code","previous-step":"onboarding-create-mesh"},null,8,["next-step"])]),_:1})]),_:1})]),_:1})}}});const D=u(L,[["__scopeId","data-v-589f21f1"]]);export{D as default};
