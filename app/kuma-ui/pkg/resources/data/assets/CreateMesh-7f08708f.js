import{d as p,i as u,r as _,j as m,o as h,k as b,w as t,a as s,e,g as o,y as g,u as n,D as v,S as f,N as y,O as x,H as S}from"./index-e4d45609.js";import{O as N,a as A,b as C}from"./OnboardingPage-bb6683d6.js";const r=a=>(y("data-v-bc48623a"),a=a(),x(),a),k={class:"mb-4 text-center"},D=r(()=>o("i",null,"default",-1)),O=r(()=>o("p",{class:"mt-4 text-center"},`
        This mesh is empty. Next, you add services and their data plane proxies.
      `,-1)),M=p({__name:"CreateMesh",setup(a){const c=[{label:"Name",key:"name"},{label:"Services",key:"servicesAmount"},{label:"DPPs",key:"dppsAmount"}],i=u(),d=_({total:1,data:[{name:"default",servicesAmount:0,dppsAmount:0}]}),l=m(()=>i.getters["config/getMulticlusterStatus"]?"onboarding-multi-zone":"onboarding-configuration-types");return(w,B)=>(h(),b(C,null,{header:t(()=>[s(N,null,{title:t(()=>[e(`
          Create the mesh
        `)]),_:1})]),content:t(()=>[o("p",k,[e(`
        When you install, `+g(n(v))+" creates a ",1),D,e(` mesh, but you can add as many meshes as you need.
      `)]),e(),s(n(f),{class:"table",fetcher:()=>d.value,headers:c,"disable-pagination":""},null,8,["fetcher"]),e(),O]),navigation:t(()=>[s(A,{"next-step":"onboarding-add-services","previous-step":n(l)},null,8,["previous-step"])]),_:1}))}});const P=S(M,[["__scopeId","data-v-bc48623a"]]);export{P as default};
