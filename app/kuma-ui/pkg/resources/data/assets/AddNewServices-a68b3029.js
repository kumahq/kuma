import{O as b,a as x,b as S}from"./OnboardingPage-3d45a474.js";import{d as m,o as u,f as h,q as w,H as y,c as k,a as B,b as t,i as s,e as _,h as o,l as e,p as $,n as C}from"./index-64044ff8.js";import{h as v,f as N,k as I,g as A,_ as O}from"./RouteView.vue_vue_type_script_setup_true_lang-75502ce3.js";import{_ as M}from"./RouteTitle.vue_vue_type_script_setup_true_lang-6fc4ecf7.js";import"./kongponents.es-b630df1e.js";const V=""+new URL("new-service-demo-bff0792e.svg",import.meta.url).href,q=""+new URL("new-service-manually-5bec5301.svg",import.meta.url).href,z=m({__name:"ServiceBox",props:{active:{type:Boolean,required:!1,default:!1}},emits:["clicked"],setup(a,{emit:n}){const c=a;return(r,i)=>(u(),h("div",{class:y(["box",{"box--active":c.active}]),"data-testid":"box",onClick:i[0]||(i[0]=l=>n("clicked"))},[w(r.$slots,"default",{},void 0,!0)],2))}});const p=v(z,[["__scopeId","data-v-93fc7d1a"]]),f=a=>($("data-v-8218f0d7"),a=a(),C(),a),H={class:"service-mode-list"},L=f(()=>e("div",{class:"service-box-content"},[e("img",{src:V}),o(),e("p",{class:"service-mode-title"},`
                  Demo app
                `),o(),e("p",null,"Counter application")],-1)),R=f(()=>e("div",{class:"service-box-content"},[e("img",{src:q}),o(),e("p",{class:"service-mode-title"},`
                  Manually
                `),o(),e("p",null,"After this wizard")],-1)),U=m({__name:"AddNewServices",setup(a){const n=N(),{t:c}=I(),r=k(()=>n.state.onboarding.mode==="manually"?"onboarding-completed":"onboarding-add-services-code");function i(l){n.dispatch("onboarding/changeMode",l)}return(l,d)=>(u(),B(O,null,{default:t(()=>[s(M,{title:_(c)("onboarding.routes.add-services.title")},null,8,["title"]),o(),s(A,null,{default:t(()=>[s(b,null,{header:t(()=>[s(x,null,{title:t(()=>[o(`
              Add services
            `)]),_:1})]),content:t(()=>[e("div",H,[s(p,{active:_(n).state.onboarding.mode==="demo",onClicked:d[0]||(d[0]=g=>i("demo"))},{default:t(()=>[L]),_:1},8,["active"]),o(),s(p,{active:_(n).state.onboarding.mode==="manually",onClicked:d[1]||(d[1]=g=>i("manually"))},{default:t(()=>[R]),_:1},8,["active"])])]),navigation:t(()=>[s(S,{"next-step":r.value,"previous-step":"onboarding-create-mesh"},null,8,["next-step"])]),_:1})]),_:1})]),_:1}))}});const F=v(U,[["__scopeId","data-v-8218f0d7"]]);export{F as default};
