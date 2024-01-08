import{K as v}from"./index-fce48c05.js";import{d as b,a as c,o as r,b as d,w as n,e as a,f as e,m as t,t as l,c as m,F as y,D as V,l as u,E as x,p as N,v as O,x as I,_ as C}from"./index-a963f507.js";import{O as k,a as R,b as A}from"./OnboardingPage-2e812b76.js";const _=s=>(O("data-v-6d3e602f"),s=s(),I(),s),E=_(()=>t("strong",null,"few minutes",-1)),S={"data-testid":"kuma-environment"},W=_(()=>t("h2",{class:"text-center"},`
              Let’s get started:
            `,-1)),z={class:"item-status-list-wrapper"},B={class:"item-status-list"},L={class:"circle mr-2"},T=b({__name:"OnboardingWelcomeView",setup(s){return(K,D)=>{const p=c("RouteTitle"),g=c("AppView"),f=c("RouteView");return r(),d(f,{name:"onboarding-welcome-view"},{default:n(({env:h,t:o,can:w})=>[a(p,{title:o("onboarding.routes.welcome.title",{name:o("common.product.name")}),render:!1},null,8,["title"]),e(),a(g,null,{default:n(()=>[t("div",null,[a(k,null,{header:n(()=>[a(R,null,{title:n(()=>[e(`
                Welcome to `+l(o("common.product.name")),1)]),description:n(()=>[t("p",null,[e(`
                  Congratulations on downloading `+l(o("common.product.name"))+"! You are just a ",1),E,e(` away from getting your service mesh fully online.
                `)]),e(),t("p",null,[e(`
                  We have automatically detected that you are running on `),t("strong",S,l(o(`common.product.environment.${h("KUMA_ENVIRONMENT")}`)),1),e(`.
                `)])]),_:2},1024)]),content:n(()=>[W,e(),t("div",z,[t("ul",B,[(r(!0),m(y,null,V([{name:`Run ${o("common.product.name")} control plane`,status:!0},{name:"Learn about deployments",status:!1},{name:"Learn about configuration storage",status:!1},...w("use zones")?[{name:"Add zones",status:!1}]:[],{name:"Create the mesh",status:!1},{name:"Add services",status:!1},{name:"Go to the dashboard",status:!1}],i=>(r(),m("li",{key:i.name},[t("span",L,[i.status?(r(),d(u(x),{key:0,size:u(v)},null,8,["size"])):N("",!0)]),e(" "+l(i.name),1)]))),128))])])]),navigation:n(()=>[a(A,{"next-step":"onboarding-deployment-types-view"})]),_:2},1024)])]),_:2},1024)]),_:1})}}});const $=C(T,[["__scopeId","data-v-6d3e602f"]]);export{$ as default};
