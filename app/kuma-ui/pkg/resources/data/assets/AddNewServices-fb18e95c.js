import{O as g,a as b,b as x}from"./OnboardingPage-c91d802d.js";import{d as _,o as m,h as S,N as w,q as h,c as y,a as k,w as s,e as c,f as a,g as e,u as l,p as B,j as N}from"./runtime-dom.esm-bundler-32659b48.js";import{_ as v}from"./_plugin-vue_export-helper-c27b6911.js";import{u as C}from"./store-15db9444.js";import"./kongponents.es-c2485d1e.js";import"./production-58f5acfb.js";const A=""+new URL("new-service-demo-bff0792e.svg",import.meta.url).href,I=""+new URL("new-service-manually-5bec5301.svg",import.meta.url).href,$=_({__name:"ServiceBox",props:{active:{type:Boolean,required:!1,default:!1}},emits:["clicked"],setup(t,{emit:o}){const r=t;return(i,n)=>(m(),S("div",{class:h(["box",{"box--active":r.active}]),"data-testid":"box",onClick:n[0]||(n[0]=d=>o("clicked"))},[w(i.$slots,"default",{},void 0,!0)],2))}});const p=v($,[["__scopeId","data-v-93fc7d1a"]]),u=t=>(B("data-v-99dd6812"),t=t(),N(),t),O={class:"service-mode-list"},M=u(()=>e("div",{class:"service-box-content"},[e("img",{src:A}),a(),e("p",{class:"service-mode-title"},`
              Demo app
            `),a(),e("p",null,"Counter application")],-1)),V=u(()=>e("div",{class:"service-box-content"},[e("img",{src:I}),a(),e("p",{class:"service-mode-title"},`
              Manually
            `),a(),e("p",null,"After this wizard")],-1)),q=_({__name:"AddNewServices",setup(t){const o=C(),r=y(()=>o.state.onboarding.mode==="manually"?"onboarding-completed":"onboarding-add-services-code");function i(n){o.dispatch("onboarding/changeMode",n)}return(n,d)=>(m(),k(g,null,{header:s(()=>[c(b,null,{title:s(()=>[a(`
          Add services
        `)]),_:1})]),content:s(()=>[e("div",O,[c(p,{active:l(o).state.onboarding.mode==="demo",onClicked:d[0]||(d[0]=f=>i("demo"))},{default:s(()=>[M]),_:1},8,["active"]),a(),c(p,{active:l(o).state.onboarding.mode==="manually",onClicked:d[1]||(d[1]=f=>i("manually"))},{default:s(()=>[V]),_:1},8,["active"])])]),navigation:s(()=>[c(x,{"next-step":l(r),"previous-step":"onboarding-create-mesh"},null,8,["next-step"])]),_:1}))}});const E=v(q,[["__scopeId","data-v-99dd6812"]]);export{E as default};
