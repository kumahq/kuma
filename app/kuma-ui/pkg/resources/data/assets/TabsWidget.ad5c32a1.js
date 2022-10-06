import{_ as g,y,G as B,u as k,x as v,e as d,A as u,k as n,f as c,c as l,B as T,h as E,w as h,r as i,o as t,a as S,p as w,l as L}from"./index.dbfc69fe.js";import{E as I}from"./ErrorBlock.f52010c3.js";import{_ as K}from"./LoadingBlock.vue_vue_type_script_setup_true_lang.f2781028.js";const V={name:"TabsWidget",components:{ErrorBlock:I,LoadingBlock:K,KIcon:y,KTabs:B},props:{loaders:{type:Boolean,default:!0},isLoading:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},error:{type:Object,required:!1,default:null},tabs:{type:Array,required:!0},hasBorder:{type:Boolean,default:!1},initialTabOverride:{type:String,default:null}},emits:["on-tab-change"],data(){return{tabState:this.initialTabOverride&&`#${this.initialTabOverride}`}},computed:{tabsSlots(){return this.tabs.map(e=>e.hash.replace("#",""))},isReady(){return this.loaders!==!1?!this.isEmpty&&!this.hasError&&!this.isLoading:!0}},methods:{switchTab(e){k.logger.info(v.TABS_TAB_CHANGE,{data:{newTab:e}}),this.$emit("on-tab-change",e)}}},x=e=>(w("data-v-951a8820"),e=e(),L(),e),A={class:"tab-container","data-testid":"tab-container"},C={key:0,class:"tab__header"},W={class:"tab__content-container"},N={class:"flex items-center with-warnings"},O=x(()=>c("span",null," Warnings ",-1)),H={key:1};function R(e,o,s,q,_,r){const m=i("KIcon"),p=i("KTabs"),b=i("LoadingBlock"),f=i("ErrorBlock");return t(),d("div",A,[e.$slots.tabHeader&&r.isReady?(t(),d("header",C,[u(e.$slots,"tabHeader",{},void 0,!0)])):n("",!0),c("div",W,[r.isReady?(t(),l(p,{key:0,modelValue:_.tabState,"onUpdate:modelValue":o[0]||(o[0]=a=>_.tabState=a),tabs:s.tabs,onChanged:o[1]||(o[1]=a=>r.switchTab(a))},T({"warnings-anchor":h(()=>[c("span",N,[S(m,{class:"mr-1",icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"16"}),O])]),_:2},[E(r.tabsSlots,a=>({name:a,fn:h(()=>[u(e.$slots,a,{},void 0,!0)])}))]),1032,["modelValue","tabs"])):n("",!0),s.loaders===!0?(t(),d("div",H,[s.isLoading?(t(),l(b,{key:0})):s.hasError?(t(),l(f,{key:1,error:s.error},null,8,["error"])):n("",!0)])):n("",!0)])])}const U=g(V,[["render",R],["__scopeId","data-v-951a8820"]]);export{U as T};
