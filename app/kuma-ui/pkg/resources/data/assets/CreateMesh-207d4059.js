import{P as m,z as p}from"./production-c33f040b.js";import{z as u}from"./kongponents.es-c2485d1e.js";import{a as _,O as g}from"./OnboardingNavigation-0c346547.js";import{O as b}from"./OnboardingPage-fe05b34b.js";import{_ as h}from"./_plugin-vue_export-helper-c27b6911.js";import{i as a,a as f,w as n,o as v,e as o,f as e,g as t,t as y}from"./runtime-dom.esm-bundler-32659b48.js";import"./store-96085224.js";const O={name:"CreateMesh",components:{OnboardingNavigation:_,OnboardingHeading:g,OnboardingPage:b,KTable:u},data(){return{productName:m,tableHeaders:[{label:"Name",key:"name"},{label:"Services",key:"servicesAmount"},{label:"DPPs",key:"dppsAmount"}],tableData:{total:1,data:[{name:"default",servicesAmount:0,dppsAmount:0}]}}},computed:{...p({multicluster:"config/getMulticlusterStatus"}),previousStep(){return this.multicluster?"onboarding-multi-zone":"onboarding-configuration-types"}}},x={class:"text-center mb-4"},N=t("i",null,"default",-1),P={class:"flex justify-center mt-10 mb-12 pb-12"},C={class:"w-full sm:w-3/5 lg:w-2/5 p-4"},T=t("p",{class:"text-center"},`
        This mesh is empty. Next, you add services and their data plane proxies.
      `,-1);function k(w,A,D,H,s,r){const i=a("OnboardingHeading"),c=a("KTable"),l=a("OnboardingNavigation"),d=a("OnboardingPage");return v(),f(d,null,{header:n(()=>[o(i,null,{title:n(()=>[e(`
          Create the mesh
        `)]),_:1})]),content:n(()=>[t("p",x,[e(`
        When you install, `+y(s.productName)+" creates a ",1),N,e(` mesh, but you can add as many meshes as you need.
      `)]),e(),t("div",P,[t("div",C,[o(c,{fetcher:()=>s.tableData,headers:s.tableHeaders,"disable-pagination":"","is-small":""},null,8,["fetcher","headers"])])]),e(),T]),navigation:n(()=>[o(l,{"next-step":"onboarding-add-services","previous-step":r.previousStep},null,8,["previous-step"])]),_:1})}const E=h(O,[["render",k]]);export{E as default};
